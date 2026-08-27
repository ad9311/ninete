package handlers

import (
	"net/http"
	"sort"

	"github.com/ad9311/ninete/internal/repo"
)

// apiDashboardSummary is GetDashboard's dashboardSummary as JSON.
// TopCategories reuses apiExpenseStatRow (handle_api_expense_stats.go): both
// are a category name paired with a total, so there is no second type for
// the same shape.
type apiDashboardSummary struct {
	ThisMonthTotal  uint64              `json:"this_month_total"`
	LastMonthTotal  uint64              `json:"last_month_total"`
	MonthChangeSign string              `json:"month_change_sign"`
	MonthChangePct  int                 `json:"month_change_pct"`
	TopCategories   []apiExpenseStatRow `json:"top_categories"`
}

type apiDashboardResponse struct {
	Data apiDashboardSummary `json:"data"`
}

// GetAPIDashboard is GetDashboard's JSON twin: explicit [start, end) bounds
// for the current and prior month, computed client-side by
// lib/dateRanges.ts, instead of tz_offset (§3.6 of docs/spa-migration.md).
func (h *Handler) GetAPIDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	q := r.URL.Query()

	thisStart, thisEnd, err := parseAPIRequiredDateBounds(q, "this_start", "this_end")
	if err != nil {
		h.WriteAPIError(w, err, ErrAPIInvalidDateRange)

		return
	}

	lastStart, lastEnd, err := parseAPIRequiredDateBounds(q, "last_start", "last_end")
	if err != nil {
		h.WriteAPIError(w, err, ErrAPIInvalidDateRange)

		return
	}

	thisFilters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: user.ID, Operator: "="},
			{Name: "date", Value: thisStart, Operator: ">="},
			{Name: "date", Value: thisEnd, Operator: "<"},
		},
		Connector: "AND",
	}
	lastFilters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: user.ID, Operator: "="},
			{Name: "date", Value: lastStart, Operator: ">="},
			{Name: "date", Value: lastEnd, Operator: "<"},
		},
		Connector: "AND",
	}

	thisTotals, err := h.store.FindExpensesCategoryTotals(ctx, thisFilters)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	lastTotals, err := h.store.FindExpensesCategoryTotals(ctx, lastFilters)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	var thisMonthTotal uint64
	for _, t := range thisTotals {
		thisMonthTotal += t.Total
	}

	var lastMonthTotal uint64
	for _, t := range lastTotals {
		lastMonthTotal += t.Total
	}

	var sign string
	var pct int
	if lastMonthTotal > 0 {
		if thisMonthTotal >= lastMonthTotal {
			sign = "+"
			pct = safeUint64ToInt((thisMonthTotal - lastMonthTotal) * 100 / lastMonthTotal)
		} else {
			sign = "-"
			pct = safeUint64ToInt((lastMonthTotal - thisMonthTotal) * 100 / lastMonthTotal)
		}
	}

	_, nameByID, err := h.findCategories(ctx)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	catRows := make([]apiExpenseStatRow, 0, len(thisTotals))
	for _, t := range thisTotals {
		catRows = append(catRows, apiExpenseStatRow{
			Name:  categoryNameOrUnknown(nameByID, t.CategoryID),
			Total: t.Total,
		})
	}
	sort.Slice(catRows, func(i, j int) bool {
		return catRows[i].Total > catRows[j].Total
	})
	if len(catRows) > 5 {
		catRows = catRows[:5]
	}

	h.WriteJSON(w, http.StatusOK, apiDashboardResponse{
		Data: apiDashboardSummary{
			ThisMonthTotal:  thisMonthTotal,
			LastMonthTotal:  lastMonthTotal,
			MonthChangeSign: sign,
			MonthChangePct:  pct,
			TopCategories:   catRows,
		},
	})
}
