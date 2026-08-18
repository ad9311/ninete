package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetExpensesBudgets(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/expenses/budgets", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_single_month_columns_for_this_month",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "budget_month", "budget_month@example.com", "budget_password_1")
				category := s.CreateCategory(t, "budget month category")
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "budget month expense", 60000, monthStart(0)))
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{category.ID: 50000})
				cookies := s.AuthCookies(t, "budget_month@example.com", "budget_password_1")

				req := spec.NewGetRequest("/expenses/budgets", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, "budget month category")
				require.Contains(t, body, "Left")
				require.Contains(t, body, "$500.00")
				// Spent 600 against a 500 budget: 120% and a negative remainder.
				require.Contains(t, body, "120%")
				require.Contains(t, body, "-$100.00")
			},
		},
		{
			name: "should_render_per_month_rows_for_six_months",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "budget_months", "budget_months@example.com", "budget_password_2")
				category := s.CreateCategory(t, "budget months category")
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "budget months expense a", 40000, monthStart(-1)))
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "budget months expense b", 70000, monthStart(-2)))
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{category.ID: 50000})
				cookies := s.AuthCookies(t, "budget_months@example.com", "budget_password_2")

				req := spec.NewGetRequest("/expenses/budgets?date_range=six_months", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, "budget months category")
				require.Contains(t, body, "months over")
				require.Contains(t, body, monthLabel(-1))
				require.Contains(t, body, monthLabel(-2))
				require.Contains(t, body, "$500.00/mo")
			},
		},
		{
			name: "should_fall_back_to_this_month_for_an_unsupported_range",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "budget_range", "budget_range@example.com", "budget_password_3")
				category := s.CreateCategory(t, "budget range category")
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{category.ID: 50000})
				cookies := s.AuthCookies(t, "budget_range@example.com", "budget_password_3")

				req := spec.NewGetRequest("/expenses/budgets?date_range=all_time", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.NotContains(t, body, `value="all_time"`)
				require.Contains(t, body, `value="this_month"`)
				// The single-month branch is what renders a Left column.
				require.Contains(t, body, "Left")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostExpensesBudgets(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	user := s.CreateAuthUser(t, "budget_post", "budget_post@example.com", "budget_password_4")
	category := s.CreateCategory(t, "budget post category")
	cookies := s.AuthCookies(t, "budget_post@example.com", "budget_password_4")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_save_a_budget_from_the_form",
			fn: func(t *testing.T) {
				csrfToken, formCookies := s.CSRFFrom(t, "/expenses/budgets", cookies)
				body := fmt.Sprintf("budget_%d=45000&date_range=this_month", category.ID)

				req := spec.NewPostRequest("/expenses/budgets", body, formCookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/expenses/budgets?date_range=this_month", rec.Header().Get("Location"))

				budgets, err := s.Store.FindExpenseBudgets(t.Context(), user.ID)
				require.NoError(t, err)
				require.Len(t, budgets, 1)
				require.Equal(t, uint64(45000), budgets[0].Amount)
			},
		},
		{
			name: "should_clear_a_budget_when_the_field_is_blank",
			fn: func(t *testing.T) {
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{category.ID: 45000})

				csrfToken, formCookies := s.CSRFFrom(t, "/expenses/budgets", cookies)
				body := fmt.Sprintf("budget_%d=&date_range=this_month", category.ID)

				req := spec.NewPostRequest("/expenses/budgets", body, formCookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)

				budgets, err := s.Store.FindExpenseBudgets(t.Context(), user.ID)
				require.NoError(t, err)
				require.Empty(t, budgets)
			},
		},
		{
			name: "should_reject_a_malformed_budget_field",
			fn: func(t *testing.T) {
				csrfToken, formCookies := s.CSRFFrom(t, "/expenses/budgets", cookies)

				req := spec.NewPostRequest("/expenses/budgets", "budget_abc=100", formCookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

// monthStart returns the first instant of the month offset months away from
// now, in UTC — the same zone computeDateRange builds its boundaries in.
func monthStart(offset int) int64 {
	now := time.Now().UTC()

	return time.Date(now.Year(), now.Month()+time.Month(offset), 1, 0, 0, 0, 0, time.UTC).Unix()
}

func monthLabel(offset int) string {
	return time.Unix(monthStart(offset), 0).UTC().Format("2006-01")
}
