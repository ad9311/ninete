package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

type apiExpenseBody struct {
	ID           int      `json:"id"`
	CategoryID   int      `json:"category_id"`
	CategoryName string   `json:"category_name"`
	Description  string   `json:"description"`
	Amount       uint64   `json:"amount"`
	Date         int64    `json:"date"`
	CreatedAt    int64    `json:"created_at"`
	Tags         []string `json:"tags"`
}

type apiExpenseListBody struct {
	Data       []apiExpenseBody `json:"data"`
	Pagination struct {
		CurrentPage int    `json:"current_page"`
		TotalPages  int    `json:"total_pages"`
		PerPage     int    `json:"per_page"`
		TotalCount  int    `json:"total_count"`
		HasPrev     bool   `json:"has_prev"`
		HasNext     bool   `json:"has_next"`
		SortField   string `json:"sort_field"`
		SortOrder   string `json:"sort_order"`
		CategoryID  int    `json:"category_id"`
	} `json:"pagination"`
}

// monthBounds returns the UTC-midnight [start, end) epoch bounds for the
// calendar month containing t, matching what the client resolves a named
// range to before calling /api/expenses* (§3.6 of docs/spa-migration.md).
func monthBounds(t time.Time) (int64, int64) {
	year, month, _ := t.UTC().Date()
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	return start.Unix(), start.AddDate(0, 1, 0).Unix()
}

func TestAPIExpenses(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_require_authentication",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/api/expenses", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "should_create_list_show_update_and_delete_an_expense",
			fn: func(t *testing.T) {
				_, cookies, csrfToken := apiUser(t, s, "api_exp_user_1", "api_exp_user_1@example.com", "api_exp_password_1")
				category := s.CreateCategory(t, "api_exp_cat_1")

				res, body := doJSON(t, handler, http.MethodPost, "/api/expenses", map[string]any{
					"category_id": category.ID,
					"description": "Groceries",
					"amount":      2599,
					"date":        1755993600, // 2025-08-24T00:00:00Z
					"tags":        []string{"Food", "food", " Weekly "},
				}, cookies, csrfToken)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var created apiExpenseBody
				require.NoError(t, json.Unmarshal(body, &created))
				require.NotZero(t, created.ID)
				require.Equal(t, "Groceries", created.Description)
				require.Equal(t, uint64(2599), created.Amount)
				require.Equal(t, int64(1755993600), created.Date)
				require.ElementsMatch(t, []string{"food", "weekly"}, created.Tags)

				res, body = doJSON(t, handler, http.MethodGet, "/api/expenses", nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)

				var list apiExpenseListBody
				require.NoError(t, json.Unmarshal(body, &list))
				require.Len(t, list.Data, 1)
				require.Equal(t, created.ID, list.Data[0].ID)
				require.Equal(t, category.Name, list.Data[0].CategoryName)
				require.Equal(t, 1, list.Pagination.TotalCount)

				showURL := "/api/expenses/" + itoa(created.ID)
				res, body = doJSON(t, handler, http.MethodGet, showURL, nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)

				var shown apiExpenseBody
				require.NoError(t, json.Unmarshal(body, &shown))
				require.Equal(t, created.ID, shown.ID)

				res, body = doJSON(t, handler, http.MethodPut, showURL, map[string]any{
					"category_id": category.ID,
					"description": "Groceries (updated)",
					"amount":      3099,
					"date":        1755993600,
					"tags":        []string{"food"},
				}, cookies, csrfToken)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var updated apiExpenseBody
				require.NoError(t, json.Unmarshal(body, &updated))
				require.Equal(t, "Groceries (updated)", updated.Description)
				require.Equal(t, uint64(3099), updated.Amount)
				require.Equal(t, []string{"food"}, updated.Tags)

				res, _ = doJSON(t, handler, http.MethodDelete, showURL, nil, cookies, csrfToken)
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				res, _ = doJSON(t, handler, http.MethodGet, showURL, nil, cookies, "")
				require.Equal(t, http.StatusNotFound, res.StatusCode)
			},
		},
		{
			name: "should_reject_missing_required_fields_with_422_and_field_rules",
			fn: func(t *testing.T) {
				_, cookies, csrfToken := apiUser(t, s, "api_exp_user_2", "api_exp_user_2@example.com", "api_exp_password_2")

				res, body := doJSON(t, handler, http.MethodPost, "/api/expenses", map[string]any{
					"description": "Missing category",
					"amount":      100,
					"date":        1755993600,
				}, cookies, csrfToken)
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				var apiErr handlers.APIError
				require.NoError(t, json.Unmarshal(body, &apiErr))
				require.Equal(t, "required", apiErr.Fields["category_id"])
			},
		},
		{
			name: "should_not_serve_another_users_expense",
			fn: func(t *testing.T) {
				ownerID, _, _ := apiUser(t, s, "api_exp_owner", "api_exp_owner@example.com", "api_exp_password_3")
				category := s.CreateCategory(t, "api_exp_cat_3")
				expense := s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID:  category.ID,
						Description: "Owner only",
						Amount:      500,
					},
					Date: 1755993600,
				})

				_, otherCookies, _ := apiUser(t, s, "api_exp_other", "api_exp_other@example.com", "api_exp_password_4")

				res, _ := doJSON(t, handler, http.MethodGet, "/api/expenses/"+itoa(expense.ID), nil, otherCookies, "")
				require.Equal(t, http.StatusNotFound, res.StatusCode)
			},
		},
		{
			name: "should_filter_by_explicit_date_bounds_category_and_search",
			fn: func(t *testing.T) {
				ownerID, cookies, _ := apiUser(t, s, "api_exp_user_5", "api_exp_user_5@example.com", "api_exp_password_5")
				categoryA := s.CreateCategory(t, "api_exp_cat_5a")
				categoryB := s.CreateCategory(t, "api_exp_cat_5b")

				inRange := s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID:  categoryA.ID,
						Description: "Coffee shop",
						Amount:      450,
					},
					Date: 1755993600, // 2025-08-24
				})
				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID:  categoryB.ID,
						Description: "Coffee beans",
						Amount:      1200,
					},
					Date: 1753488000, // 2025-07-26, a different month
				})

				start, end := monthBounds(time.Unix(1755993600, 0))
				url := "/api/expenses?start=" + itoa64(start) + "&end=" + itoa64(end) + "&category_id=" + itoa(categoryA.ID)

				res, body := doJSON(t, handler, http.MethodGet, url, nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)

				var list apiExpenseListBody
				require.NoError(t, json.Unmarshal(body, &list))
				require.Len(t, list.Data, 1)
				require.Equal(t, inRange.ID, list.Data[0].ID)

				res, body = doJSON(t, handler, http.MethodGet, "/api/expenses?q=beans", nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)
				require.NoError(t, json.Unmarshal(body, &list))
				require.Len(t, list.Data, 1)
				require.Equal(t, "Coffee beans", list.Data[0].Description)
			},
		},
		{
			name: "should_reject_an_invalid_date_range",
			fn: func(t *testing.T) {
				_, cookies, _ := apiUser(t, s, "api_exp_user_6", "api_exp_user_6@example.com", "api_exp_password_6")

				res, _ := doJSON(t, handler, http.MethodGet, "/api/expenses?start=200&end=100", nil, cookies, "")
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
			},
		},
		{
			name: "should_get_expense_stats_sorted_by_total",
			fn: func(t *testing.T) {
				ownerID, cookies, _ := apiUser(t, s, "api_exp_user_7", "api_exp_user_7@example.com", "api_exp_password_7")
				categoryA := s.CreateCategory(t, "api_exp_cat_7a")
				categoryB := s.CreateCategory(t, "api_exp_cat_7b")

				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{CategoryID: categoryA.ID, Description: "Small", Amount: 100},
					Date:              1755993600,
				})
				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{CategoryID: categoryB.ID, Description: "Big", Amount: 900},
					Date:              1755993600,
				})

				res, body := doJSON(t, handler, http.MethodGet, "/api/expenses/stats", nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)

				var stats struct {
					Data []struct {
						Name  string `json:"name"`
						Total uint64 `json:"total"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &stats))
				require.Len(t, stats.Data, 2)
				// Default sort is total DESC.
				require.Equal(t, categoryB.Name, stats.Data[0].Name)
				require.Equal(t, uint64(900), stats.Data[0].Total)
			},
		},
		{
			name: "should_get_and_update_expense_budgets",
			fn: func(t *testing.T) {
				ownerID, cookies, csrfToken := apiUser(
					t, s, "api_exp_user_8", "api_exp_user_8@example.com", "api_exp_password_8",
				)
				category := s.CreateCategory(t, "api_exp_cat_8")

				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{CategoryID: category.ID, Description: "Rent", Amount: 5000},
					Date:              1755993600,
				})
				s.SaveExpenseBudgets(t, ownerID, map[int]uint64{category.ID: 10000})

				start, end := monthBounds(time.Unix(1755993600, 0))
				url := "/api/expenses/budgets?start=" + itoa64(start) + "&end=" + itoa64(end) + "&mode=month"

				res, body := doJSON(t, handler, http.MethodGet, url, nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)

				var budgets struct {
					Mode string `json:"mode"`
					Rows []struct {
						CategoryName string `json:"category_name"`
						Total        uint64 `json:"total"`
						Budget       uint64 `json:"budget"`
						Left         int64  `json:"left"`
					} `json:"rows"`
					EditRows []struct {
						CategoryID int    `json:"category_id"`
						Amount     uint64 `json:"amount"`
					} `json:"edit_rows"`
				}
				require.NoError(t, json.Unmarshal(body, &budgets))
				require.Equal(t, "month", budgets.Mode)
				require.Len(t, budgets.Rows, 1)
				require.Equal(t, uint64(5000), budgets.Rows[0].Total)
				require.Equal(t, uint64(10000), budgets.Rows[0].Budget)
				require.Equal(t, int64(5000), budgets.Rows[0].Left)

				res, _ = doJSON(t, handler, http.MethodPut, "/api/expenses/budgets", map[string]any{
					"amounts": map[string]any{itoa(category.ID): 0},
				}, cookies, csrfToken)
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				res, body = doJSON(t, handler, http.MethodGet, url, nil, cookies, "")
				require.Equal(t, http.StatusOK, res.StatusCode)
				require.NoError(t, json.Unmarshal(body, &budgets))
				require.False(t, budgets.Rows[0].Budget > 0)
			},
		},
		{
			name: "should_reject_a_write_without_a_csrf_token",
			fn: func(t *testing.T) {
				_, cookies, _ := apiUser(t, s, "api_exp_user_9", "api_exp_user_9@example.com", "api_exp_password_9")
				category := s.CreateCategory(t, "api_exp_cat_9")

				res, _ := doJSON(t, handler, http.MethodPost, "/api/expenses", map[string]any{
					"category_id": category.ID,
					"description": "No token",
					"amount":      100,
					"date":        1755993600,
				}, cookies, "")

				require.Equal(t, http.StatusForbidden, res.StatusCode)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

func TestAPIExpensesQuick(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_ask_for_a_category_the_first_time_then_remember_it",
			fn: func(t *testing.T) {
				_, cookies, csrfToken := apiUser(
					t, s, "api_quick_user_1", "api_quick_user_1@example.com", "api_quick_password_1",
				)
				category := s.CreateCategory(t, "api_quick_cat_1")

				res, body := doJSON(t, handler, http.MethodPost, "/api/expenses/quick", map[string]any{
					"quick_input": "Uber, 12.50, today",
					"tz_offset":   0,
				}, cookies, csrfToken)
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				var apiErr handlers.APIError
				require.NoError(t, json.Unmarshal(body, &apiErr))
				require.Equal(t, "required", apiErr.Fields["category_id"])

				res, body = doJSON(t, handler, http.MethodPost, "/api/expenses/quick", map[string]any{
					"quick_input": "Uber, 12.50, today",
					"category_id": category.ID,
					"tz_offset":   0,
				}, cookies, csrfToken)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var created apiExpenseBody
				require.NoError(t, json.Unmarshal(body, &created))
				require.Equal(t, "Uber", created.Description)
				require.Equal(t, uint64(1250), created.Amount)

				// Same description again: the remembered mapping resolves the
				// category without the client sending one.
				res, body = doJSON(t, handler, http.MethodPost, "/api/expenses/quick", map[string]any{
					"quick_input": "Uber, 8.00, today",
					"tz_offset":   0,
				}, cookies, csrfToken)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var second apiExpenseBody
				require.NoError(t, json.Unmarshal(body, &second))
				require.Equal(t, category.Name, second.CategoryName)
			},
		},
		{
			name: "should_reject_a_malformed_quick_input",
			fn: func(t *testing.T) {
				_, cookies, csrfToken := apiUser(
					t, s, "api_quick_user_2", "api_quick_user_2@example.com", "api_quick_password_2",
				)

				res, _ := doJSON(t, handler, http.MethodPost, "/api/expenses/quick", map[string]any{
					"quick_input": "not enough fields",
					"tz_offset":   0,
				}, cookies, csrfToken)
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

func itoa64(v int64) string {
	return itoa(int(v))
}
