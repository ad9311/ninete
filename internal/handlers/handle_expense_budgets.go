package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

const budgetFieldPrefix = "budget_"

// budgetMonthLayout matches strftime('%Y-%m') in selectExpensesCategoryMonthTotals.
const budgetMonthLayout = "2006-01"

// budgetRow is one category on the budgets page. Months, MonthsOver,
// MonthCount and AvgPerMonth are filled in budgetModeMonths only.
type budgetRow struct {
	CategoryName string
	Total        uint64
	HasBudget    bool
	Budget       uint64
	Left         int64
	Pct          int
	BarPct       int
	Over         bool
	Months       []budgetMonthRow
	MonthsOver   int
	MonthCount   int
	AvgPerMonth  uint64
}

type budgetMonthRow struct {
	Month  string
	Total  uint64
	Pct    int
	BarPct int
	Over   bool
}

// budgetEditRow is one row of the edit form, which lists every category so a
// category with neither budget nor spend can still be given one.
type budgetEditRow struct {
	CategoryID int
	Name       string
	Amount     uint64
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

func (h *Handler) GetExpensesBudgets(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)

	rangeKey, mode, rows, editRows, ok := h.buildBudgetsPage(w, r)
	if !ok {
		return
	}

	data["budgetMode"] = string(mode)
	data["dateRange"] = rangeKey
	data["dateRanges"] = budgetDateRanges
	data["rows"] = rows
	data["editRows"] = editRows

	h.render(w, http.StatusOK, ExpensesBudgets, data)
}

func (h *Handler) PostExpensesBudgets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	amountByCategoryID, err := parseExpenseBudgetsForm(r)
	if err != nil {
		h.renderBudgetsErr(w, r, err)

		return
	}

	if err := h.store.SaveExpenseBudgets(ctx, user.ID, amountByCategoryID); err != nil {
		h.renderBudgetsErr(w, r, err)

		return
	}

	target := "/expenses/budgets"
	if rangeKey := r.FormValue("date_range"); rangeKey != "" {
		normalized, _ := budgetDateRange(rangeKey)
		target += "?date_range=" + normalized
	}

	http.Redirect(w, r, target, http.StatusSeeOther)
}

// ----------------------------------------------------------------------------- //
// Unexported Functions and Helpers
// ----------------------------------------------------------------------------- //

// buildBudgetsPage loads everything the budgets page renders. It reports false
// once it has written an error response of its own.
func (h *Handler) buildBudgetsPage(
	w http.ResponseWriter,
	r *http.Request,
) (string, budgetMode, []budgetRow, []budgetEditRow, bool) {
	ctx := r.Context()
	user := getCurrentUser(r)

	rangeKey, mode := budgetDateRange(budgetRangeKey(r))

	filters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: user.ID, Operator: "="},
		},
		Connector: "AND",
	}

	dr, ok := computeDateRange(rangeKey, parseTZOffset(r))
	if !ok {
		h.renderErr(w, r, http.StatusInternalServerError, ExpensesBudgets, ErrUnknownDateRange)

		return "", "", nil, nil, false
	}

	filters.FilterFields = append(filters.FilterFields,
		repo.FilterField{Name: "date", Value: dr.start, Operator: ">="},
		repo.FilterField{Name: "date", Value: dr.end, Operator: "<"},
	)

	monthTotals, err := h.store.FindExpensesCategoryMonthTotals(ctx, filters)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ExpensesBudgets, err)

		return "", "", nil, nil, false
	}

	budgets, err := h.store.FindExpenseBudgets(ctx, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ExpensesBudgets, err)

		return "", "", nil, nil, false
	}

	categories, categoryNameByID, ok := h.findCategoriesOrErr(w, r, ExpensesBudgets)
	if !ok {
		return "", "", nil, nil, false
	}

	budgetByCategoryID := make(map[int]uint64, len(budgets))
	for _, b := range budgets {
		budgetByCategoryID[b.CategoryID] = b.Amount
	}

	rows := buildBudgetRows(monthTotals, budgetByCategoryID, categoryNameByID, mode, budgetMonths(dr, time.Now()))
	editRows := buildBudgetEditRows(categories, budgetByCategoryID)

	return rangeKey, mode, rows, editRows, true
}

// renderBudgetsErr re-renders the page with the form error shown. The page data
// is rebuilt from scratch: the failed submission changed nothing.
func (h *Handler) renderBudgetsErr(w http.ResponseWriter, r *http.Request, err error) {
	rangeKey, mode, rows, editRows, ok := h.buildBudgetsPage(w, r)
	if !ok {
		return
	}

	data := h.tmplData(r)
	data["budgetMode"] = string(mode)
	data["dateRange"] = rangeKey
	data["dateRanges"] = budgetDateRanges
	data["rows"] = rows
	data["editRows"] = editRows
	data["error"] = err.Error()

	h.render(w, http.StatusBadRequest, ExpensesBudgets, data)
}

// budgetRangeKey reads the requested range. A GET carries it in the query; the
// POST form posts it in the body, and the failed-submission re-render has to
// find it there or the page snaps back to this_month.
func budgetRangeKey(r *http.Request) string {
	if key := r.URL.Query().Get("date_range"); key != "" {
		return key
	}

	return r.FormValue("date_range")
}

// budgetMonths lists the calendar months the range covers, oldest first, keyed
// the way SelectExpensesCategoryMonthTotals groups them. It is the denominator
// of the "N of M months over" summary, so months with no spending count: a
// month under budget because nothing was bought is still under budget.
//
// The end is clamped to the current month. this_year always spans to January of
// the next year, and dividing a part-finished year by twelve would deflate the
// average with months that have not happened yet.
func budgetMonths(dr dateRange, now time.Time) []string {
	start := time.Unix(dr.start, 0).UTC()
	// dr.end is exclusive, so the last month in range is the one before it.
	last := time.Unix(dr.end, 0).UTC().AddDate(0, 0, -1)

	if nowUTC := now.UTC(); nowUTC.Before(last) {
		last = nowUTC
	}

	months := make([]string, 0, 12)
	for m := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC); !m.After(last); m = m.AddDate(0, 1, 0) {
		months = append(months, m.Format(budgetMonthLayout))
	}

	if len(months) == 0 {
		months = append(months, start.Format(budgetMonthLayout))
	}

	return months
}

func buildBudgetRows(
	monthTotals []repo.ExpenseCategoryMonthTotal,
	budgetByCategoryID map[int]uint64,
	categoryNameByID map[int]string,
	mode budgetMode,
	monthKeys []string,
) []budgetRow {
	totalByCategoryID := make(map[int]uint64, len(monthTotals))
	monthsByCategoryID := make(map[int][]repo.ExpenseCategoryMonthTotal, len(monthTotals))

	for _, t := range monthTotals {
		totalByCategoryID[t.CategoryID] += t.Total
		monthsByCategoryID[t.CategoryID] = append(monthsByCategoryID[t.CategoryID], t)
	}

	monthKeys = withSpentMonths(monthKeys, monthTotals)
	monthCount := len(monthKeys)

	categoryIDs := make([]int, 0, len(totalByCategoryID)+len(budgetByCategoryID))
	for categoryID := range totalByCategoryID {
		categoryIDs = append(categoryIDs, categoryID)
	}
	for categoryID := range budgetByCategoryID {
		if _, seen := totalByCategoryID[categoryID]; !seen {
			categoryIDs = append(categoryIDs, categoryID)
		}
	}

	rows := make([]budgetRow, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		budget := budgetByCategoryID[categoryID]
		total := totalByCategoryID[categoryID]

		row := budgetRow{
			CategoryName: categoryNameOrUnknown(categoryNameByID, categoryID),
			Total:        total,
			HasBudget:    budget > 0,
			Budget:       budget,
			MonthCount:   monthCount,
		}

		if row.HasBudget {
			row.Left = budgetLeft(budget, total)
			row.Pct, row.BarPct = budgetPercent(total, budget)

			if mode == budgetModeMonths {
				row.Months, row.MonthsOver = buildBudgetMonthRows(monthsByCategoryID[categoryID], monthKeys, budget)
				row.AvgPerMonth = total / uint64(monthCount) //nolint:gosec // budgetMonths returns at least one month
				// Budget is a monthly amount, so a multi-month total is not
				// comparable to it — $700 spent over six months against a $500
				// monthly budget is well under. Only a month that individually
				// exceeded the budget makes the row over.
				row.Over = row.MonthsOver > 0
			} else {
				row.Over = total > budget
			}
		}

		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].CategoryName < rows[j].CategoryName
	})

	return rows
}

// withSpentMonths adds any month that carries spending but falls outside the
// clamped month list, so no expense is counted in a row total without a bar to
// account for it. An expense may be dated ahead of today — a purchase made now
// can be billed next month — which puts it past the clamp.
func withSpentMonths(monthKeys []string, totals []repo.ExpenseCategoryMonthTotal) []string {
	known := make(map[string]bool, len(monthKeys))
	for _, key := range monthKeys {
		known[key] = true
	}

	extended := monthKeys
	for _, t := range totals {
		if !known[t.Month] {
			known[t.Month] = true
			extended = append(extended, t.Month)
		}
	}

	if len(extended) == len(monthKeys) {
		return monthKeys
	}

	sort.Strings(extended)

	return extended
}

// buildBudgetMonthRows renders one bar per month in range, not per month that
// happened to have spending. A month with nothing spent is under budget and has
// to appear, or the list contradicts the "N of M months" count beside it.
func buildBudgetMonthRows(
	totals []repo.ExpenseCategoryMonthTotal,
	monthKeys []string,
	budget uint64,
) ([]budgetMonthRow, int) {
	totalByMonth := make(map[string]uint64, len(totals))
	for _, t := range totals {
		totalByMonth[t.Month] += t.Total
	}

	months := make([]budgetMonthRow, 0, len(monthKeys))
	over := 0

	for _, key := range monthKeys {
		total := totalByMonth[key]
		pct, barPct := budgetPercent(total, budget)
		month := budgetMonthRow{
			Month:  key,
			Total:  total,
			Pct:    pct,
			BarPct: barPct,
			Over:   total > budget,
		}

		if month.Over {
			over++
		}

		months = append(months, month)
	}

	return months, over
}

// budgetLeft is the signed remainder of a budget. Both operands are cent
// amounts one person entered by hand, so neither half of the subtraction can
// approach the int64 range.
func budgetLeft(budget, total uint64) int64 {
	if budget >= total {
		return int64(budget - total) //nolint:gosec // cent amount, far below int64 max
	}

	return -int64(total - budget) //nolint:gosec // cent amount, far below int64 max
}

// budgetPercent returns the true percent and the percent clamped to 100 for the
// <progress> element. A zero budget never reaches here — a cleared amount is
// deleted rather than stored — but it is guarded anyway, since it would divide.
func budgetPercent(total, budget uint64) (int, int) {
	if budget == 0 {
		return 0, 0
	}

	pct := int(total * 100 / budget) //nolint:gosec // both operands are page-sized cent amounts
	if pct > 100 {
		return pct, 100
	}

	return pct, pct
}

func buildBudgetEditRows(categories []repo.Category, budgetByCategoryID map[int]uint64) []budgetEditRow {
	rows := make([]budgetEditRow, 0, len(categories))
	for _, category := range categories {
		rows = append(rows, budgetEditRow{
			CategoryID: category.ID,
			Name:       category.Name,
			Amount:     budgetByCategoryID[category.ID],
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Name < rows[j].Name
	})

	return rows
}

// parseExpenseBudgetsForm reads the budget_<categoryID> fields. An empty field
// means no budget and arrives as zero, which SaveExpenseBudgets deletes.
func parseExpenseBudgetsForm(r *http.Request) (map[int]uint64, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseForm, err)
	}

	amountByCategoryID := make(map[int]uint64)

	for field, values := range r.Form {
		if !strings.HasPrefix(field, budgetFieldPrefix) {
			continue
		}

		categoryID, err := strconv.Atoi(strings.TrimPrefix(field, budgetFieldPrefix))
		if err != nil || categoryID < 1 {
			return nil, ErrBudgetCategoryField
		}

		raw := ""
		if len(values) > 0 {
			raw = strings.TrimSpace(values[0])
		}

		if raw == "" {
			amountByCategoryID[categoryID] = 0

			continue
		}

		amount, err := prog.ParseAmount(raw)
		if err != nil {
			return nil, err
		}

		amountByCategoryID[categoryID] = amount
	}

	return amountByCategoryID, nil
}
