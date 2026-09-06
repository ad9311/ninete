package logic

import (
	"context"
	"sort"
	"time"

	"github.com/ad9311/ninete/internal/repo"
)

// UntaggedSectionName labels the residual section holding every expense that
// carries none of the configured grouping tags. It is pinned last rather than
// sorted with the others, being a leftover bucket and not a real group
// (docs/monthly-report.md, "Grouping rules").
const UntaggedSectionName = "Untagged"

// ReportExpense is one line of the report's expense list.
//
// There is no billed date on it: every expense in a report shares one, by
// definition of the period, and it is already printed in the header.
type ReportExpense struct {
	Description  string
	CategoryName string
	Amount       uint64
}

// ReportSection is one tag's expenses, or the residual Untagged bucket.
type ReportSection struct {
	Name     string
	Total    uint64
	Expenses []ReportExpense
	Untagged bool
}

// ReportCategoryTotal is one bar of the report's category chart.
type ReportCategoryTotal struct {
	Name  string
	Total uint64
}

// ReportBudget is one category's budget line for the reported month. It is the
// single-month shape of the budgets page: a monthly budget compared against
// one month of spending needs no averaging.
//
// There is no clamped bar percent here, unlike the budgets endpoint's row: the
// clamp exists to keep a <progress> element from overflowing, and the PDF's
// budget table prints figures rather than drawing bars.
type ReportBudget struct {
	CategoryName string
	Total        uint64
	Budget       uint64
	Left         int64
	Pct          int
	Over         bool
}

// MonthlyReport is everything the PDF renders, already ordered. Assembling it
// here rather than in the renderer is what lets the layout be tested against a
// literal and the arithmetic against the database, separately.
type MonthlyReport struct {
	// Month is the UTC-midnight first day of the reported month, the same
	// value the expenses in it carry as their billed date.
	Month          time.Time
	Total          uint64
	ExpenseCount   int
	Sections       []ReportSection
	CategoryTotals []ReportCategoryTotal
	Budgets        []ReportBudget
	// Grouped is false when the settings name no tags, in which case Sections
	// holds one unnamed section and the PDF prints a flat list.
	Grouped bool
}

// MonthBounds returns the [start, end) epoch-second bounds of the calendar
// month containing month.
//
// Both are UTC midnights because that is what "expenses"."date" stores: the
// billed date is month-precision and carries no zone, so unlike the instant
// columns the expense search filters, no client zone is needed to turn a month
// into a window (docs/monthly-report.md, "Timezone").
func MonthBounds(month time.Time) (int64, int64) {
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	return start.Unix(), start.AddDate(0, 1, 0).Unix()
}

// BuildMonthlyReport assembles the report for the calendar month containing
// month, using the user's saved grouping tags.
func (s *Store) BuildMonthlyReport(ctx context.Context, userID int, month time.Time) (MonthlyReport, error) {
	start, end := MonthBounds(month)

	report := MonthlyReport{Month: time.Unix(start, 0).UTC()}

	setting, err := s.FindReportSetting(ctx, userID)
	if err != nil {
		return report, err
	}

	// Read once and threaded through everything below. Categories are a
	// shared lookup table, so every part of the report resolves an id through
	// the same map — which is also what makes the budget table's name keying
	// safe: both sides of that lookup come from here.
	categoryNameByID, err := s.reportCategoryNames(ctx)
	if err != nil {
		return report, err
	}

	expenses, err := s.reportExpenses(ctx, userID, start, end)
	if err != nil {
		return report, err
	}

	sectionByExpenseID, err := s.reportSectionNames(ctx, userID, start, end, setting.TagIDs)
	if err != nil {
		return report, err
	}

	report.Grouped = len(setting.TagIDs) > 0
	report.Sections = buildReportSections(expenses, categoryNameByID, sectionByExpenseID, report.Grouped)
	report.ExpenseCount = len(expenses)

	for _, expense := range expenses {
		report.Total += expense.Amount
	}

	if report.CategoryTotals, err = s.reportCategoryTotals(ctx, userID, start, end, categoryNameByID); err != nil {
		return report, err
	}

	if report.Budgets, err = s.reportBudgets(ctx, userID, report.CategoryTotals, categoryNameByID); err != nil {
		return report, err
	}

	return report, nil
}

// reportCategoryNames reads the shared category lookup table once per report.
func (s *Store) reportCategoryNames(ctx context.Context) (map[int]string, error) {
	categories, err := s.queries.SelectCategories(ctx)
	if err != nil {
		return nil, err
	}

	nameByID := make(map[int]string, len(categories))
	for _, category := range categories {
		nameByID[category.ID] = category.Name
	}

	return nameByID, nil
}

func (s *Store) reportExpenses(
	ctx context.Context,
	userID int,
	start, end int64,
) ([]repo.Expense, error) {
	expenses, err := s.queries.SelectExpenses(ctx, repo.QueryOptions{
		// Largest first, so a section reads as a ranking rather than as the
		// order the rows happened to be entered in.
		Sorting: repo.Sorting{Field: "amount", Order: "DESC"},
		Filters: repo.Filters{
			FilterFields: []repo.FilterField{
				{Name: "user_id", Value: userID, Operator: "="},
				{Name: "date", Value: start, Operator: ">="},
				{Name: "date", Value: end, Operator: "<"},
			},
			Connector: "AND",
		},
	})
	if err != nil {
		return nil, err
	}

	return expenses, nil
}

// reportSectionNames maps each expense to the grouping tag that owns it. The
// query returns every matching tagging in tie-break order, so the first row
// seen for an expense is the one that decides — later rows for the same
// expense are the other grouping tags it also carries, and are dropped.
func (s *Store) reportSectionNames(
	ctx context.Context,
	userID int,
	start, end int64,
	tagIDs []int,
) (map[int]string, error) {
	rows, err := s.queries.SelectExpenseGroupTags(ctx, userID, start, end, tagIDs)
	if err != nil {
		return nil, err
	}

	nameByExpenseID := make(map[int]string, len(rows))
	for _, row := range rows {
		if _, taken := nameByExpenseID[row.ExpenseID]; taken {
			continue
		}

		nameByExpenseID[row.ExpenseID] = row.TagName
	}

	return nameByExpenseID, nil
}

func (s *Store) reportCategoryTotals(
	ctx context.Context,
	userID int,
	start, end int64,
	categoryNameByID map[int]string,
) ([]ReportCategoryTotal, error) {
	totals, err := s.queries.SelectExpensesCategoryTotals(ctx, repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: userID, Operator: "="},
			{Name: "date", Value: start, Operator: ">="},
			{Name: "date", Value: end, Operator: "<"},
		},
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}

	out := make([]ReportCategoryTotal, 0, len(totals))
	for _, total := range totals {
		out = append(out, ReportCategoryTotal{
			Name:  reportCategoryName(categoryNameByID, total.CategoryID),
			Total: total.Total,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Total != out[j].Total {
			return out[i].Total > out[j].Total
		}

		return out[i].Name < out[j].Name
	})

	return out, nil
}

// reportBudgets compares each budgeted category against the month's spending.
// Categories are keyed by name because the totals arrive already resolved to
// names, and a category with a budget but no spending still gets a row — a
// budget kept to zero spend is a result worth seeing.
func (s *Store) reportBudgets(
	ctx context.Context,
	userID int,
	categoryTotals []ReportCategoryTotal,
	nameByID map[int]string,
) ([]ReportBudget, error) {
	budgets, err := s.FindExpenseBudgets(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(budgets) == 0 {
		return nil, nil
	}

	totalByName := make(map[string]uint64, len(categoryTotals))
	for _, total := range categoryTotals {
		totalByName[total.Name] = total.Total
	}

	out := make([]ReportBudget, 0, len(budgets))
	for _, budget := range budgets {
		name := reportCategoryName(nameByID, budget.CategoryID)
		total := totalByName[name]

		pct, _ := BudgetPercent(total, budget.Amount)
		out = append(out, ReportBudget{
			CategoryName: name,
			Total:        total,
			Budget:       budget.Amount,
			Left:         BudgetLeft(budget.Amount, total),
			Pct:          pct,
			Over:         total > budget.Amount,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CategoryName < out[j].CategoryName
	})

	return out, nil
}

// sectionKey identifies a section by what it is, not by what it is called. The
// residual bucket is a different section from a tag that happens to share its
// label, so the two must not collide in the lookup below.
type sectionKey struct {
	name     string
	untagged bool
}

// buildReportSections assigns every expense to its section and orders the
// result: by total descending, with Untagged pinned last and empty sections
// simply never created, since a section only exists once something lands in it.
//
// Whether an expense is residual is decided by the assignment — no grouping
// tag was resolved for it — and never by comparing the resulting name against
// UntaggedSectionName. A user's tags cannot collide with that label today,
// since every tag name is lowercased by prog.NormalizeLowerTrim, but a
// structural fact should not rest on a normalization rule two packages away.
func buildReportSections(
	expenses []repo.Expense,
	categoryNameByID map[int]string,
	sectionByExpenseID map[int]string,
	grouped bool,
) []ReportSection {
	byKey := make(map[sectionKey]*ReportSection)
	order := make([]*ReportSection, 0)

	for _, expense := range expenses {
		name := ""
		untagged := false

		if grouped {
			name = sectionByExpenseID[expense.ID]
			if name == "" {
				name, untagged = UntaggedSectionName, true
			}
		}

		key := sectionKey{name: name, untagged: untagged}

		section, ok := byKey[key]
		if !ok {
			section = &ReportSection{Name: name, Untagged: untagged}
			byKey[key] = section
			order = append(order, section)
		}

		section.Total += expense.Amount
		section.Expenses = append(section.Expenses, ReportExpense{
			Description:  expense.Description,
			CategoryName: reportCategoryName(categoryNameByID, expense.CategoryID),
			Amount:       expense.Amount,
		})
	}

	sort.SliceStable(order, func(i, j int) bool {
		if order[i].Untagged != order[j].Untagged {
			return order[j].Untagged
		}

		return order[i].Total > order[j].Total
	})

	out := make([]ReportSection, 0, len(order))
	for _, section := range order {
		out = append(out, *section)
	}

	return out
}

// reportCategoryName keeps a deleted or unknown category from printing as an
// empty cell. Categories are a shared lookup table, so this should not happen;
// a blank column would hide it if it did.
func reportCategoryName(nameByID map[int]string, categoryID int) string {
	if name := nameByID[categoryID]; name != "" {
		return name
	}

	return "Uncategorized"
}
