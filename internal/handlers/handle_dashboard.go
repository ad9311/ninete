package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"sort"

	"github.com/ad9311/ninete/internal/repo"
)

type dashboardSummary struct {
	ThisMonthTotal  uint64
	LastMonthTotal  uint64
	MonthChangeSign string
	MonthChangePct  int
	TopCategories   []expenseCategoryRow
}

type dashboardMacros struct {
	HasGoal       bool
	Goal          repo.MacroGoal
	TodayTotals   repo.MacroDayTotals
	TodayProgress macroProgressData
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	user := getCurrentUser(r)

	summary, ok := h.buildDashboardSummary(w, r, user.ID)
	if !ok {
		return
	}

	macros, ok := h.buildDashboardMacros(w, r, user.ID, r.URL.Query().Get("date"))
	if !ok {
		return
	}

	data["summary"] = summary
	data["macros"] = macros

	h.render(w, http.StatusOK, DashboardIndex, data)
}

func (h *Handler) buildDashboardSummary(w http.ResponseWriter, r *http.Request, userID int) (dashboardSummary, bool) {
	ctx := r.Context()

	tzOffset := parseTZOffset(r)
	thisDR, _ := computeDateRange("this_month", tzOffset)
	lastDR, _ := computeDateRange("last_month", tzOffset)

	thisFilters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: userID, Operator: "="},
			{Name: "date", Value: thisDR.start, Operator: ">="},
			{Name: "date", Value: thisDR.end, Operator: "<"},
		},
		Connector: "AND",
	}
	lastFilters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: userID, Operator: "="},
			{Name: "date", Value: lastDR.start, Operator: ">="},
			{Name: "date", Value: lastDR.end, Operator: "<"},
		},
		Connector: "AND",
	}

	thisTotals, err := h.store.FindExpensesCategoryTotals(ctx, thisFilters)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, DashboardIndex, err)

		return dashboardSummary{}, false
	}

	lastTotals, err := h.store.FindExpensesCategoryTotals(ctx, lastFilters)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, DashboardIndex, err)

		return dashboardSummary{}, false
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

	_, nameByID, ok := h.findCategoriesOrErr(w, r, DashboardIndex)
	if !ok {
		return dashboardSummary{}, false
	}

	catRows := make([]expenseCategoryRow, 0, len(thisTotals))
	for _, t := range thisTotals {
		catRows = append(catRows, expenseCategoryRow{
			CategoryName: categoryNameOrUnknown(nameByID, t.CategoryID),
			Total:        t.Total,
		})
	}
	sort.Slice(catRows, func(i, j int) bool {
		return catRows[i].Total > catRows[j].Total
	})
	if len(catRows) > 5 {
		catRows = catRows[:5]
	}

	return dashboardSummary{
		ThisMonthTotal:  thisMonthTotal,
		LastMonthTotal:  lastMonthTotal,
		MonthChangeSign: sign,
		MonthChangePct:  pct,
		TopCategories:   catRows,
	}, true
}

func (h *Handler) buildDashboardMacros(
	w http.ResponseWriter, r *http.Request, userID int, dateStr string,
) (dashboardMacros, bool) {
	ctx := r.Context()

	goal, err := h.store.FindMacroGoal(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return dashboardMacros{}, true
	}
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, DashboardIndex, err)

		return dashboardMacros{}, false
	}

	dayStart, nextDay, _ := computeDayWindow(dateStr)

	todayTotals, err := h.store.FindMacroDayTotals(ctx, userID, dayStart, nextDay, "")
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, DashboardIndex, err)

		return dashboardMacros{}, false
	}

	return dashboardMacros{
		HasGoal:       true,
		Goal:          goal,
		TodayTotals:   todayTotals,
		TodayProgress: computeMacroProgress(todayTotals, goal),
	}, true
}
