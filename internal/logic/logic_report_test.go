package logic_test

import (
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestBuildMonthlyReport(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	// A fixed past month rather than "now", so the assertions do not move with
	// the clock.
	monthlyReportMonth := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)

	user := s.CreateAuthUser(t, "report_build_user", "report_build_user@example.com", "report_build_pass")
	food := s.CreateCategory(t, "report build food")
	travel := s.CreateCategory(t, "report build travel")

	monthStart := monthlyReportMonth.Unix()

	create := func(t *testing.T, description string, amount uint64, categoryID int, tags []string) {
		t.Helper()

		s.CreateExpense(t, user.ID, logic.ExpenseParams{
			Description: description,
			Amount:      amount,
			Date:        monthStart,
			CategoryID:  categoryID,
			Tags:        tags,
		})
	}

	cardTag := s.CreateTag(t, user.ID, "rep_card")
	transferTag := s.CreateTag(t, user.ID, "rep_transfer")

	create(t, "report groceries", 5000, food.ID, []string{"rep_card"})
	create(t, "report flight", 30000, travel.ID, []string{"rep_transfer"})
	create(t, "report snack", 1000, food.ID, nil)

	// Billed to the following month, so nothing in the report may count it.
	s.CreateExpense(t, user.ID, logic.ExpenseParams{
		Description: "report next month",
		Amount:      99999,
		Date:        monthlyReportMonth.AddDate(0, 1, 0).Unix(),
		CategoryID:  food.ID,
	})

	saveTags := func(t *testing.T, tagIDs []int) {
		t.Helper()

		require.NoError(t, s.Store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
			Timezone: "UTC",
			TagIDs:   tagIDs,
		}))
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_cover_only_the_reported_month",
			fn: func(t *testing.T) {
				saveTags(t, nil)

				report, err := s.Store.BuildMonthlyReport(ctx, user.ID, monthlyReportMonth)
				require.NoError(t, err)

				require.Equal(t, 3, report.ExpenseCount)
				require.Equal(t, uint64(36000), report.Total)
				require.Equal(t, monthlyReportMonth, report.Month)
			},
		},
		{
			name: "should_print_one_flat_section_with_no_grouping_tags",
			fn: func(t *testing.T) {
				saveTags(t, nil)

				report, err := s.Store.BuildMonthlyReport(ctx, user.ID, monthlyReportMonth)
				require.NoError(t, err)

				require.False(t, report.Grouped)
				require.Len(t, report.Sections, 1)
				require.Empty(t, report.Sections[0].Name)
				require.Len(t, report.Sections[0].Expenses, 3)
			},
		},
		{
			name: "should_group_by_tag_ordered_by_total_with_untagged_last",
			fn: func(t *testing.T) {
				saveTags(t, []int{cardTag.ID, transferTag.ID})

				report, err := s.Store.BuildMonthlyReport(ctx, user.ID, monthlyReportMonth)
				require.NoError(t, err)

				require.True(t, report.Grouped)
				require.Len(t, report.Sections, 3)

				// $300 transfer outranks the $50 card section, and Untagged is
				// pinned last despite carrying more than nothing.
				require.Equal(t, transferTag.Name, report.Sections[0].Name)
				require.Equal(t, uint64(30000), report.Sections[0].Total)
				require.Equal(t, cardTag.Name, report.Sections[1].Name)
				require.Equal(t, uint64(5000), report.Sections[1].Total)
				require.Equal(t, logic.UntaggedSectionName, report.Sections[2].Name)
				require.True(t, report.Sections[2].Untagged)
				require.Equal(t, uint64(1000), report.Sections[2].Total)
			},
		},
		{
			name: "should_omit_a_grouping_tag_with_nothing_in_the_month",
			fn: func(t *testing.T) {
				empty := s.CreateTag(t, user.ID, "rep_empty")
				saveTags(t, []int{cardTag.ID, empty.ID})

				report, err := s.Store.BuildMonthlyReport(ctx, user.ID, monthlyReportMonth)
				require.NoError(t, err)

				for _, section := range report.Sections {
					require.NotEqual(t, empty.Name, section.Name,
						"an empty section reached the report")
				}
			},
		},
		{
			name: "should_count_a_multiply_tagged_expense_once",
			fn: func(t *testing.T) {
				// Both taggings are written by one call, so they share a
				// created_at to the second and only the id separates them —
				// which is exactly the tie-break the grouping rule relies on.
				// The first name given is inserted first, so it wins.
				create(t, "report both tags", 2000, food.ID,
					[]string{cardTag.Name, transferTag.Name})

				saveTags(t, []int{cardTag.ID, transferTag.ID})

				report, err := s.Store.BuildMonthlyReport(ctx, user.ID, monthlyReportMonth)
				require.NoError(t, err)

				seen := 0
				sectioned := uint64(0)
				landedIn := ""

				for _, section := range report.Sections {
					sectioned += section.Total

					for _, expense := range section.Expenses {
						if expense.Description == "report both tags" {
							seen++
							landedIn = section.Name
						}
					}
				}

				require.Equal(t, 1, seen, "expense appeared in more than one section")
				require.Equal(t, cardTag.Name, landedIn,
					"the earlier tagging did not decide the section")
				require.Equal(t, report.Total, sectioned,
					"section subtotals do not sum to the report total")
			},
		},
		{
			name: "should_rank_category_totals_by_amount",
			fn: func(t *testing.T) {
				report, err := s.Store.BuildMonthlyReport(ctx, user.ID, monthlyReportMonth)
				require.NoError(t, err)

				require.NotEmpty(t, report.CategoryTotals)
				require.Equal(t, travel.Name, report.CategoryTotals[0].Name)

				for i := 1; i < len(report.CategoryTotals); i++ {
					require.GreaterOrEqual(t,
						report.CategoryTotals[i-1].Total, report.CategoryTotals[i].Total)
				}
			},
		},
		{
			name: "should_compare_budgets_against_the_month",
			fn: func(t *testing.T) {
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{
					travel.ID: 10000,
					food.ID:   100000,
				})

				report, err := s.Store.BuildMonthlyReport(ctx, user.ID, monthlyReportMonth)
				require.NoError(t, err)

				byName := make(map[string]logic.ReportBudget, len(report.Budgets))
				for _, budget := range report.Budgets {
					byName[budget.CategoryName] = budget
				}

				travelBudget := byName[travel.Name]
				require.True(t, travelBudget.Over)
				require.Equal(t, uint64(30000), travelBudget.Total)
				require.Equal(t, int64(-20000), travelBudget.Left)
				require.Equal(t, 300, travelBudget.Pct,
					"the printed percent is the true one, not clamped")

				foodBudget := byName[food.Name]
				require.False(t, foodBudget.Over)
				require.Positive(t, foodBudget.Left)
			},
		},
		{
			name: "should_not_leak_another_users_expenses",
			fn: func(t *testing.T) {
				stranger := s.CreateAuthUser(
					t, "report_stranger", "report_stranger@example.com", "report_stranger_pass",
				)
				s.CreateExpense(t, stranger.ID, logic.ExpenseParams{
					Description: "report stranger expense",
					Amount:      777,
					Date:        monthStart,
					CategoryID:  food.ID,
				})

				report, err := s.Store.BuildMonthlyReport(ctx, user.ID, monthlyReportMonth)
				require.NoError(t, err)

				for _, section := range report.Sections {
					for _, expense := range section.Expenses {
						require.NotEqual(t, "report stranger expense", expense.Description)
					}
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

func TestMonthBounds(t *testing.T) {
	start, end := logic.MonthBounds(time.Date(2026, time.December, 17, 23, 30, 0, 0, time.UTC))

	require.Equal(t, time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC).Unix(), start)
	require.Equal(t, time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC).Unix(), end,
		"the year must roll over")
}
