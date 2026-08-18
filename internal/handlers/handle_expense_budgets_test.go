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
			// Reproduction: Over was computed as total > budget in both modes,
			// so a range total spread under a monthly budget flagged the row.
			name: "should_not_flag_a_multi_month_row_whose_every_month_stayed_under",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "budget_under", "budget_under@example.com", "budget_password_5")
				category := s.CreateCategory(t, "budget under category")
				// 60000 total against a 50000 monthly budget, but 30000 in each
				// of two months: never over in any single month.
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "budget under a", 30000, monthStart(-1)))
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "budget under b", 30000, monthStart(-2)))
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{category.ID: 50000})
				cookies := s.AuthCookies(t, "budget_under@example.com", "budget_password_5")

				req := spec.NewGetRequest("/expenses/budgets?date_range=six_months", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, "budget under category")
				require.Contains(t, body, "0 of 6 months over")
				require.NotContains(t, body, "budget-row-over")
			},
		},
		{
			// Reproduction: the denominator came from month-6 through month+1,
			// which is seven months under a "Last 6 months" label.
			name: "should_count_exactly_six_months_for_six_months",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "budget_six", "budget_six@example.com", "budget_password_6")
				category := s.CreateCategory(t, "budget six category")
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{category.ID: 50000})
				cookies := s.AuthCookies(t, "budget_six@example.com", "budget_password_6")

				req := spec.NewGetRequest("/expenses/budgets?date_range=six_months", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, "0 of 6 months over")
				// Every month in range gets a bar, including the empty ones.
				for offset := 0; offset > -6; offset-- {
					require.Contains(t, body, monthLabel(offset))
				}
				require.NotContains(t, body, monthLabel(-6))
			},
		},
		{
			// Reproduction: this_year always spans to January of the next year,
			// so the denominator was 12 and the average was divided by twelve
			// even in a year still months from finishing.
			name: "should_count_only_elapsed_months_for_this_year",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "budget_year", "budget_year@example.com", "budget_password_7")
				category := s.CreateCategory(t, "budget year category")
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{category.ID: 50000})
				cookies := s.AuthCookies(t, "budget_year@example.com", "budget_password_7")

				req := spec.NewGetRequest("/expenses/budgets?date_range=this_year", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				elapsed := int(time.Now().UTC().Month())
				require.Contains(t, rec.Body.String(), fmt.Sprintf("0 of %d months over", elapsed))
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
		{
			// Reproduction: the re-render read the range from the query string,
			// where a POST never carries it, so a rejected submission dropped
			// the user back onto this_month.
			name: "should_keep_the_submitted_range_when_the_form_is_rejected",
			fn: func(t *testing.T) {
				csrfToken, formCookies := s.CSRFFrom(t, "/expenses/budgets", cookies)

				req := spec.NewPostRequest(
					"/expenses/budgets",
					"budget_abc=100&date_range=six_months",
					formCookies,
					csrfToken,
				)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusBadRequest, rec.Code)
				body := rec.Body.String()
				require.Contains(t, body, `name="date_range" value="six_months"`)
				// The multi-month branch renders no Left column.
				require.NotContains(t, body, "<th>Left</th>")
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
