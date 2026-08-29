package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestExportsExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			// Not a standalone reproduction, and worth being exact about why:
			// this passes even with the route deleted, because "/*" then
			// serves the SPA shell and AuthMiddleware redirects that. What it
			// pins is the pair — this case says the path redirects when
			// signed out, the next says the same path really serves the
			// export when signed in. Only both together rule out the bug,
			// which was the export answering an expired session with the API
			// chain's 401: no Location, nothing for a navigation to follow,
			// so the browser saved the JSON error envelope as expenses.json.
			// Move the route back under /api and the second case fails.
			name: "should_redirect_a_signed_out_visitor_to_the_login_page",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest(handlers.ExportExpensesPath, nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, handlers.AppLoginPath, rec.Header().Get("Location"))
				require.NotContains(t, rec.Header().Get("Content-Disposition"), "attachment")
			},
		},
		{
			name: "should_return_json_with_expense_payload",
			fn: func(t *testing.T) {
				ownerID, cookies, _ := apiUser(
					t, s, "api_exp_dl_1", "api_exp_dl_1@example.com", "api_exp_dl_password_1",
				)
				category := s.CreateCategory(t, "api_exp_dl_cat_1")
				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID:  category.ID,
						Description: "lunch",
						Amount:      1250,
					},
					Date: 1735689600,
					Tags: []string{"food"},
				})

				res, body := doJSON(t, handler, http.MethodGet, handlers.ExportExpensesPath, nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)
				require.Contains(t, res.Header.Get("Content-Disposition"), "attachment")
				require.Contains(t, res.Header.Get("Content-Disposition"), "expenses-")

				var payload struct {
					ExportedAt int64 `json:"exported_at"`
					Expenses   []struct {
						Description string   `json:"description"`
						Tags        []string `json:"tags"`
					} `json:"expenses"`
				}
				require.NoError(t, json.Unmarshal(body, &payload))
				require.Positive(t, payload.ExportedAt)
				require.Len(t, payload.Expenses, 1)
				require.Equal(t, "lunch", payload.Expenses[0].Description)
				require.Equal(t, []string{"food"}, payload.Expenses[0].Tags)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
