package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetExports(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/exports", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_exports_page_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_idx_1", "exp_idx_1@example.com", "exp_password_1")
				cookies := s.AuthCookies(t, "exp_idx_1@example.com", "exp_password_1")

				req := spec.NewGetRequest("/exports", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestGetExportsExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/exports/expenses.json", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_json_with_expense_payload",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(t, "exp_dl_1", "exp_dl_1@example.com", "exp_password_1")
				category := s.CreateCategory(t, "exp_cat_1")
				s.CreateExpense(t, user.ID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID:  category.ID,
						Description: "lunch",
						Amount:      1250,
					},
					Date: 1735689600,
					Tags: []string{"food"},
				})
				cookies := s.AuthCookies(t, "exp_dl_1@example.com", "exp_password_1")

				req := spec.NewGetRequest("/exports/expenses.json", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
				require.Contains(t, rec.Header().Get("Content-Disposition"), "attachment")
				require.Contains(t, rec.Header().Get("Content-Disposition"), "expenses-")

				var payload struct {
					ExportedAt int64 `json:"exported_at"`
					Expenses   []struct {
						ID          int      `json:"id"`
						Description string   `json:"description"`
						Tags        []string `json:"tags"`
						Category    *struct {
							Name string `json:"name"`
						} `json:"category"`
					} `json:"expenses"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
				require.Positive(t, payload.ExportedAt)
				require.Len(t, payload.Expenses, 1)
				require.Equal(t, "lunch", payload.Expenses[0].Description)
				require.Equal(t, []string{"food"}, payload.Expenses[0].Tags)
				require.NotNil(t, payload.Expenses[0].Category)
				require.Equal(t, "exp_cat_1", payload.Expenses[0].Category.Name)
			},
		},
		{
			name: "should_exclude_other_users_expenses",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "exp_dl_2", "exp_dl_2@example.com", "exp_password_2")
				otherUser := s.CreateAuthUser(t, "exp_dl_3", "exp_dl_3@example.com", "exp_password_3")
				category := s.CreateCategory(t, "exp_cat_2")
				s.CreateExpense(t, otherUser.ID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID:  category.ID,
						Description: "private_expense",
						Amount:      500,
					},
					Date: 1735689600,
				})
				cookies := s.AuthCookies(t, "exp_dl_2@example.com", "exp_password_2")

				req := spec.NewGetRequest("/exports/expenses.json", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.NotContains(t, rec.Body.String(), "private_expense")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
