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

	rangeKey, mode := budgetDateRange(r.URL.Query().Get("date_range"))

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

	rows := buildBudgetRows(monthTotals, budgetByCategoryID, categoryNameByID, mode, monthsInRange(dr))
	editRows := buildBudgetEditRows(categories, budgetByCategoryID)

	return rangeKey, mode, rows, editRows, true
}

// renderBudgetsErr re-renders the page with the form error shown. The page data
// is rebuilt from scratch: the failed submission changed nothing.
func (h *Handler) renderBudgetsErr(w http.ResponseWriter, r *http.Request, err error) {
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

	h.renderErr(w, r, http.StatusBadRequest, ExpensesBudgets, err)
}

// monthsInRange counts the calendar months the range covers, which is the
// denominator of the "N of M months over" summary. Months with no spending
// count: a month under budget because nothing was bought is still under budget.
func monthsInRange(dr dateRange) int {
	start := time.Unix(dr.start, 0).UTC()
	end := time.Unix(dr.end, 0).UTC()

	months := (end.Year()-start.Year())*12 + int(end.Month()) - int(start.Month())
	if months < 1 {
		return 1
	}

	return months
}

func buildBudgetRows(
	monthTotals []repo.ExpenseCategoryMonthTotal,
	budgetByCategoryID map[int]uint64,
	categoryNameByID map[int]string,
	mode budgetMode,
	monthCount int,
) []budgetRow {
	totalByCategoryID := make(map[int]uint64, len(monthTotals))
	monthsByCategoryID := make(map[int][]repo.ExpenseCategoryMonthTotal, len(monthTotals))

	for _, t := range monthTotals {
		totalByCategoryID[t.CategoryID] += t.Total
		monthsByCategoryID[t.CategoryID] = append(monthsByCategoryID[t.CategoryID], t)
	}

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
			row.Over = total > budget

			if mode == budgetModeMonths {
				row.Months, row.MonthsOver = buildBudgetMonthRows(monthsByCategoryID[categoryID], budget)
				row.AvgPerMonth = total / uint64(monthCount) //nolint:gosec // monthsInRange returns at least 1
			}
		}

		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].CategoryName < rows[j].CategoryName
	})

	return rows
}

func buildBudgetMonthRows(
	totals []repo.ExpenseCategoryMonthTotal,
	budget uint64,
) ([]budgetMonthRow, int) {
	months := make([]budgetMonthRow, 0, len(totals))
	over := 0

	for _, t := range totals {
		pct, barPct := budgetPercent(t.Total, budget)
		month := budgetMonthRow{
			Month:  t.Month,
			Total:  t.Total,
			Pct:    pct,
			BarPct: barPct,
			Over:   t.Total > budget,
		}

		if month.Over {
			over++
		}

		months = append(months, month)
	}

	sort.Slice(months, func(i, j int) bool {
		return months[i].Month < months[j].Month
	})

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
