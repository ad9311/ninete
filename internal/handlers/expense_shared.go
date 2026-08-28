package handlers

import (
	"context"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/ad9311/ninete/internal/repo"
)

type dateRange struct {
	start int64
	end   int64
}

// budgetMode names how /api/expenses/budgets renders a range: one bar per
// category for a single month, or a per-month breakdown for a multi-month
// span. The client sends it explicitly (§3.6 of docs/spa-migration.md) since
// the API only ever sees the resolved bounds, not the named range key a mode
// used to be derived from server-side.
type budgetMode string

const (
	budgetModeMonth  budgetMode = "month"
	budgetModeMonths budgetMode = "months"
)

// parseAPIBoundPair reads one [start, end) pair from the two named query
// params. Neither present means "no bound" (the all_time case); either
// present alone, or start on or after end, is malformed input rather than a
// silent fallback.
func parseAPIBoundPair(q url.Values, startKey, endKey string) (start, end int64, hasBounds bool, err error) {
	rawStart, rawEnd := q.Get(startKey), q.Get(endKey)
	if rawStart == "" && rawEnd == "" {
		return 0, 0, false, nil
	}

	start, startErr := strconv.ParseInt(rawStart, 10, 64)
	end, endErr := strconv.ParseInt(rawEnd, 10, 64)
	if startErr != nil || endErr != nil || start >= end {
		return 0, 0, false, ErrAPIInvalidDateRange
	}

	return start, end, true, nil
}

// parseAPIDateBounds reads the explicit [start, end) bounds an /api/expenses*
// request carries instead of date_range+tz_offset (§3.6 of
// docs/spa-migration.md, "Retiring tz_offset on the API side"): the client
// already knows its own zone, so it resolves the named range to UTC-midnight
// epoch bounds before the fetch.
func parseAPIDateBounds(q url.Values) (start, end int64, hasBounds bool, err error) {
	return parseAPIBoundPair(q, "start", "end")
}

// parseAPIRequiredDateBounds is parseAPIBoundPair for an endpoint with no
// all_time case — /api/dashboard always compares two specific months, so a
// missing bound is malformed input rather than "no filter".
func parseAPIRequiredDateBounds(q url.Values, startKey, endKey string) (start, end int64, err error) {
	start, end, hasBounds, err := parseAPIBoundPair(q, startKey, endKey)
	if err != nil {
		return 0, 0, err
	}
	if !hasBounds {
		return 0, 0, ErrAPIInvalidDateRange
	}

	return start, end, nil
}

// monthOverMonthChange is the dashboard's "+12% vs last month" figure. An
// empty sign means there is nothing to compare against — last month recorded
// no spending — which the client shows as "No data for last month" rather
// than as a 100% jump.
func monthOverMonthChange(thisMonthTotal, lastMonthTotal uint64) (sign string, pct int) {
	if lastMonthTotal == 0 {
		return "", 0
	}

	if thisMonthTotal >= lastMonthTotal {
		return "+", safeUint64ToInt((thisMonthTotal - lastMonthTotal) * 100 / lastMonthTotal)
	}

	return "-", safeUint64ToInt((lastMonthTotal - thisMonthTotal) * 100 / lastMonthTotal)
}

func (h *Handler) findCategories(
	ctx context.Context,
) ([]repo.Category, map[int]string, error) {
	categories, err := h.store.FindCategories(ctx)
	if err != nil {
		return categories, nil, err
	}

	categoryNameByID := make(map[int]string, len(categories))
	for _, category := range categories {
		categoryNameByID[category.ID] = category.Name
	}

	return categories, categoryNameByID, nil
}

func categoryNameOrUnknown(nameByID map[int]string, categoryID int) string {
	if name := nameByID[categoryID]; name != "" {
		return name
	}

	return "Unknown"
}

func getExpense(r *http.Request) *repo.Expense {
	expense, ok := r.Context().Value(KeyExpense).(*repo.Expense)

	if !ok {
		panic("failed to get expense context")
	}

	return expense
}

func getRecurrentExpense(r *http.Request) *repo.RecurrentExpense {
	recurrentExpense, ok := r.Context().Value(KeyRecurrentExpense).(*repo.RecurrentExpense)

	if !ok {
		panic("failed to get recurrent expense context")
	}

	return recurrentExpense
}

// ----------------------------------------------------------------------------- //
// Budgets — shared by GetAPIExpenseBudgets and PutAPIExpenseBudgets
// ----------------------------------------------------------------------------- //

// budgetMonthLayout matches strftime('%Y-%m') in selectExpensesCategoryMonthTotals.
const budgetMonthLayout = "2006-01"

// budgetRow is one category on the budgets response. Months, MonthsOver,
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

// budgetPercent returns the true percent and the percent clamped to 100 for
// the client's progress bar. A zero budget never reaches here — a cleared
// amount is deleted rather than stored — but it is guarded anyway, since it
// would divide.
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
