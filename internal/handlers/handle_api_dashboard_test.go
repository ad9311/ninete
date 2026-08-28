package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

type apiDashboardBody struct {
	Data struct {
		ThisMonthTotal  uint64 `json:"this_month_total"`
		LastMonthTotal  uint64 `json:"last_month_total"`
		MonthChangeSign string `json:"month_change_sign"`
		MonthChangePct  int    `json:"month_change_pct"`
		TopCategories   []struct {
			Name  string `json:"name"`
			Total uint64 `json:"total"`
		} `json:"top_categories"`
	} `json:"data"`
}

// prevMonthBounds returns the [start, end) bounds of the calendar month before
// the one containing t. It steps back one day from the first of t's month
// rather than calling t.AddDate(0, -1, 0): Go normalizes an overflowed day, so
// on 2026-03-31 "one month earlier" is February 31 → March 3, and the two
// windows would come back identical — every expense would land in both.
func prevMonthBounds(t time.Time) (int64, int64) {
	start, _ := monthBounds(t)

	return monthBounds(time.Unix(start, 0).UTC().AddDate(0, 0, -1))
}

func dashboardURL(thisStart, thisEnd, lastStart, lastEnd int64) string {
	return "/api/dashboard?this_start=" + itoa64(thisStart) + "&this_end=" + itoa64(thisEnd) +
		"&last_start=" + itoa64(lastStart) + "&last_end=" + itoa64(lastEnd)
}

func TestAPIDashboard(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_require_authentication",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest(dashboardURL(0, 1, 0, 1), nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "should_reject_missing_bounds",
			fn: func(t *testing.T) {
				_, cookies, _ := apiUser(t, s, "api_dash_user_1", "api_dash_user_1@example.com", "api_dash_password_1")

				res, _ := doJSON(t, handler, http.MethodGet, "/api/dashboard", nil, cookies, "")

				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
			},
		},
		{
			name: "should_summarize_this_and_last_month_with_top_categories",
			fn: func(t *testing.T) {
				ownerID, cookies, _ := apiUser(
					t, s, "api_dash_user_2", "api_dash_user_2@example.com", "api_dash_password_2",
				)
				categoryA := s.CreateCategory(t, "api_dash_cat_2a")
				categoryB := s.CreateCategory(t, "api_dash_cat_2b")

				now := time.Now().UTC()
				thisStart, thisEnd := monthBounds(now)
				lastStart, lastEnd := prevMonthBounds(now)

				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{CategoryID: categoryA.ID, Description: "Rent", Amount: 3000},
					Date:              thisStart,
				})
				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{CategoryID: categoryB.ID, Description: "Food", Amount: 1000},
					Date:              thisStart,
				})
				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{CategoryID: categoryA.ID, Description: "Old rent", Amount: 2000},
					Date:              lastStart,
				})

				res, body := doJSON(
					t, handler, http.MethodGet,
					dashboardURL(thisStart, thisEnd, lastStart, lastEnd),
					nil, cookies, "",
				)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var dash apiDashboardBody
				require.NoError(t, json.Unmarshal(body, &dash))

				require.Equal(t, uint64(4000), dash.Data.ThisMonthTotal)
				require.Equal(t, uint64(2000), dash.Data.LastMonthTotal)
				require.Equal(t, "+", dash.Data.MonthChangeSign)
				require.Equal(t, 100, dash.Data.MonthChangePct)
				require.Len(t, dash.Data.TopCategories, 2)
				require.Equal(t, categoryA.Name, dash.Data.TopCategories[0].Name)
				require.Equal(t, uint64(3000), dash.Data.TopCategories[0].Total)
			},
		},
		{
			name: "should_report_no_change_when_last_month_has_no_expenses",
			fn: func(t *testing.T) {
				ownerID, cookies, _ := apiUser(
					t, s, "api_dash_user_3", "api_dash_user_3@example.com", "api_dash_password_3",
				)
				category := s.CreateCategory(t, "api_dash_cat_3")

				now := time.Now().UTC()
				thisStart, thisEnd := monthBounds(now)
				lastStart, lastEnd := prevMonthBounds(now)

				s.CreateExpense(t, ownerID, logic.ExpenseParams{
					ExpenseBaseParams: logic.ExpenseBaseParams{
						CategoryID: category.ID, Description: "Only this month", Amount: 500,
					},
					Date: thisStart,
				})

				res, body := doJSON(
					t, handler, http.MethodGet,
					dashboardURL(thisStart, thisEnd, lastStart, lastEnd),
					nil, cookies, "",
				)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var dash apiDashboardBody
				require.NoError(t, json.Unmarshal(body, &dash))

				require.Equal(t, uint64(500), dash.Data.ThisMonthTotal)
				require.Equal(t, uint64(0), dash.Data.LastMonthTotal)
				require.Empty(t, dash.Data.MonthChangeSign)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
