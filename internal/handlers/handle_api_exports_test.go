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

func TestAPIExportsExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_require_authentication",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/api/exports/expenses.json", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusUnauthorized, rec.Code)
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

				res, body := doJSON(t, handler, http.MethodGet, "/api/exports/expenses.json", nil, cookies, "")
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
