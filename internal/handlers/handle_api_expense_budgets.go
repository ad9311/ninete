package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ad9311/ninete/internal/repo"
)

type apiBudgetMonthRow struct {
	Month  string `json:"month"`
	Total  uint64 `json:"total"`
	Pct    int    `json:"pct"`
	BarPct int    `json:"bar_pct"`
	Over   bool   `json:"over"`
}

type apiBudgetRow struct {
	CategoryName string              `json:"category_name"`
	Total        uint64              `json:"total"`
	HasBudget    bool                `json:"has_budget"`
	Budget       uint64              `json:"budget"`
	Left         int64               `json:"left"`
	Pct          int                 `json:"pct"`
	BarPct       int                 `json:"bar_pct"`
	Over         bool                `json:"over"`
	Months       []apiBudgetMonthRow `json:"months,omitempty"`
	MonthsOver   int                 `json:"months_over"`
	MonthCount   int                 `json:"month_count"`
	AvgPerMonth  uint64              `json:"avg_per_month"`
}

func newAPIBudgetRow(row budgetRow) apiBudgetRow {
	months := make([]apiBudgetMonthRow, 0, len(row.Months))
	for _, m := range row.Months {
		months = append(months, apiBudgetMonthRow(m))
	}

	return apiBudgetRow{
		CategoryName: row.CategoryName,
		Total:        row.Total,
		HasBudget:    row.HasBudget,
		Budget:       row.Budget,
		Left:         row.Left,
		Pct:          row.Pct,
		BarPct:       row.BarPct,
		Over:         row.Over,
		Months:       months,
		MonthsOver:   row.MonthsOver,
		MonthCount:   row.MonthCount,
		AvgPerMonth:  row.AvgPerMonth,
	}
}

type apiBudgetEditRow struct {
	CategoryID int    `json:"category_id"`
	Name       string `json:"name"`
	Amount     uint64 `json:"amount"`
}

type apiExpenseBudgetsResponse struct {
	Mode     string             `json:"mode"`
	Rows     []apiBudgetRow     `json:"rows"`
	EditRows []apiBudgetEditRow `json:"edit_rows"`
}

// GetAPIExpenseBudgets is buildBudgetsPage's JSON twin. mode is sent by the
// client rather than derived from a date_range key server-side, because the
// API no longer receives the key — only the bounds it resolves to (§3.6 of
// docs/spa-migration.md) — and mode is exactly the piece of client-side
// knowledge (from budgetDateRanges' Value→Mode table, ported to the client)
// that bounds alone cannot recover.
func (h *Handler) GetAPIExpenseBudgets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	q := r.URL.Query()

	start, end, hasBounds, err := parseAPIDateBounds(q)
	if err != nil || !hasBounds {
		h.WriteAPIError(w, ErrAPIInvalidDateRange, ErrAPIInvalidDateRange)

		return
	}

	mode := budgetMode(q.Get("mode"))
	if mode != budgetModeMonth && mode != budgetModeMonths {
		mode = budgetModeMonth
	}

	filters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: user.ID, Operator: "="},
			{Name: "date", Value: start, Operator: ">="},
			{Name: "date", Value: end, Operator: "<"},
		},
		Connector: "AND",
	}

	monthTotals, err := h.store.FindExpensesCategoryMonthTotals(ctx, filters)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	budgets, err := h.store.FindExpenseBudgets(ctx, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	categories, categoryNameByID, err := h.findCategories(ctx)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	budgetByCategoryID := make(map[int]uint64, len(budgets))
	for _, b := range budgets {
		budgetByCategoryID[b.CategoryID] = b.Amount
	}

	rows := buildBudgetRows(
		monthTotals, budgetByCategoryID, categoryNameByID, mode, budgetMonths(dateRange{start, end}, time.Now()),
	)
	editRows := buildBudgetEditRows(categories, budgetByCategoryID)

	apiRows := make([]apiBudgetRow, 0, len(rows))
	for _, row := range rows {
		apiRows = append(apiRows, newAPIBudgetRow(row))
	}

	apiEditRows := make([]apiBudgetEditRow, 0, len(editRows))
	for _, row := range editRows {
		apiEditRows = append(apiEditRows, apiBudgetEditRow(row))
	}

	h.WriteJSON(w, http.StatusOK, apiExpenseBudgetsResponse{
		Mode:     string(mode),
		Rows:     apiRows,
		EditRows: apiEditRows,
	})
}

type expenseBudgetsRequestBody struct {
	Amounts map[string]uint64 `json:"amounts"`
}

// PutAPIExpenseBudgets is PostExpensesBudgets's JSON twin: a zero amount
// clears that category's budget, matching SaveExpenseBudgets's existing "zero
// deletes" contract. An *omitted* category is left untouched rather than
// cleared — SaveExpenseBudgets only visits the keys it is handed — so the
// client posts every edit row, the way the template form posts every field.
func (h *Handler) PutAPIExpenseBudgets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	var body expenseBudgetsRequestBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	amountByCategoryID := make(map[int]uint64, len(body.Amounts))
	for key, amount := range body.Amounts {
		categoryID, err := strconv.Atoi(key)
		if err != nil || categoryID < 1 {
			h.WriteAPIError(w, ErrBudgetCategoryField, ErrBudgetCategoryField)

			return
		}

		amountByCategoryID[categoryID] = amount
	}

	if err := h.store.SaveExpenseBudgets(ctx, user.ID, amountByCategoryID); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
