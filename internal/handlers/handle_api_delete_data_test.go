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

type apiAccountDataCountsBody struct {
	Data struct {
		Expenses          int `json:"expenses"`
		RecurrentExpenses int `json:"recurrent_expenses"`
		ExpenseBudgets    int `json:"expense_budgets"`
		Tags              int `json:"tags"`
	} `json:"data"`
}

func TestAPIDeleteData(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_require_authentication",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/api/delete-data", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "should_report_per_section_counts",
			fn: func(t *testing.T) {
				ownerID, cookies, _ := apiUser(
					t, s, "api_del_user_1", "api_del_user_1@example.com", "api_del_password_1",
				)
				category := s.CreateCategory(t, "api_del_cat_1")

				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID: category.ID, Description: "Rent", Amount: 1000,
					},
					Date: 1735689600,
				})
				s.CreateRecurrentExpense(t, ownerID, logic.RecurrentExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID: category.ID, Description: "Subscription", Amount: 500,
					},
					Period: 1,
				})
				require.NoError(t, s.Store.SaveExpenseBudgets(t.Context(), ownerID, map[int]uint64{category.ID: 2000}))
				s.CreateTag(t, ownerID, "api_del_tag_1")

				res, body := doJSON(t, handler, http.MethodGet, "/api/delete-data", nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)

				var counts apiAccountDataCountsBody
				require.NoError(t, json.Unmarshal(body, &counts))

				require.Equal(t, 1, counts.Data.Expenses)
				require.Equal(t, 1, counts.Data.RecurrentExpenses)
				require.Equal(t, 1, counts.Data.ExpenseBudgets)
				// Only the standalone tag created above: neither record was
				// given tags, so nothing else contributes to the count.
				require.Equal(t, 1, counts.Data.Tags)
			},
		},
		{
			name: "should_delete_only_the_named_section",
			fn: func(t *testing.T) {
				ownerID, cookies, csrfToken := apiUser(
					t, s, "api_del_user_2", "api_del_user_2@example.com", "api_del_password_2",
				)
				category := s.CreateCategory(t, "api_del_cat_2")

				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID: category.ID, Description: "Groceries", Amount: 300,
					},
					Date: 1735689600,
				})
				s.CreateTag(t, ownerID, "api_del_tag_2")

				res, _ := doJSON(t, handler, http.MethodDelete, "/api/delete-data/expenses", nil, cookies, csrfToken)
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				countsRes, body := doJSON(t, handler, http.MethodGet, "/api/delete-data", nil, cookies, "")
				require.Equal(t, http.StatusOK, countsRes.StatusCode)

				var counts apiAccountDataCountsBody
				require.NoError(t, json.Unmarshal(body, &counts))

				require.Zero(t, counts.Data.Expenses)
				// Deleting expenses does not remove the user's tags.
				require.Equal(t, 1, counts.Data.Tags)
			},
		},
		{
			name: "should_delete_every_section",
			fn: func(t *testing.T) {
				ownerID, cookies, csrfToken := apiUser(
					t, s, "api_del_user_3", "api_del_user_3@example.com", "api_del_password_3",
				)
				category := s.CreateCategory(t, "api_del_cat_3")

				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID: category.ID, Description: "Utilities", Amount: 400,
					},
					Date: 1735689600,
				})
				s.CreateRecurrentExpense(t, ownerID, logic.RecurrentExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID: category.ID, Description: "Gym", Amount: 200,
					},
					Period: 1,
				})
				require.NoError(t, s.Store.SaveExpenseBudgets(t.Context(), ownerID, map[int]uint64{category.ID: 1500}))
				s.CreateTag(t, ownerID, "api_del_tag_3")

				res, _ := doJSON(t, handler, http.MethodDelete, "/api/delete-data", nil, cookies, csrfToken)
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				countsRes, body := doJSON(t, handler, http.MethodGet, "/api/delete-data", nil, cookies, "")
				require.Equal(t, http.StatusOK, countsRes.StatusCode)

				var counts apiAccountDataCountsBody
				require.NoError(t, json.Unmarshal(body, &counts))

				require.Zero(t, counts.Data.Expenses)
				require.Zero(t, counts.Data.RecurrentExpenses)
				require.Zero(t, counts.Data.ExpenseBudgets)
				require.Zero(t, counts.Data.Tags)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
