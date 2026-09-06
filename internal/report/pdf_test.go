package report_test

import (
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/report"
	"github.com/stretchr/testify/require"
)

// pageCountRE reads the page total out of the document's /Pages node, which is
// the only assertion available without a PDF parser — and enough to prove the
// automatic page break fired.
var pageCountRE = regexp.MustCompile(`/Count (\d+)`)

func sampleReport() logic.MonthlyReport {
	return logic.MonthlyReport{
		Month:        time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
		Total:        36000,
		ExpenseCount: 3,
		Grouped:      true,
		Sections: []logic.ReportSection{
			{
				Name:  "transfer",
				Total: 30000,
				Expenses: []logic.ReportExpense{
					{Description: "Flight", CategoryName: "Travel", Amount: 30000},
				},
			},
			{
				Name:  "card",
				Total: 5000,
				Expenses: []logic.ReportExpense{
					{Description: "Groceries", CategoryName: "Food", Amount: 5000},
				},
			},
			{
				Name:     logic.UntaggedSectionName,
				Total:    1000,
				Untagged: true,
				Expenses: []logic.ReportExpense{
					{Description: "Snack", CategoryName: "Food", Amount: 1000},
				},
			},
		},
		CategoryTotals: []logic.ReportCategoryTotal{
			{Name: "Travel", Total: 30000},
			{Name: "Food", Total: 6000},
		},
		Budgets: []logic.ReportBudget{
			{CategoryName: "Travel", Total: 30000, Budget: 10000, Left: -20000, Pct: 300, BarPct: 100, Over: true},
			{CategoryName: "Food", Total: 6000, Budget: 100000, Left: 94000, Pct: 6, BarPct: 6},
		},
	}
}

func TestRender(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_produce_a_pdf",
			fn: func(t *testing.T) {
				out, err := report.Render(sampleReport())
				require.NoError(t, err)
				require.Greater(t, len(out), 1000, "the document is implausibly small")
				require.Equal(t, "%PDF", string(out[:4]))
				require.Contains(t, string(out[len(out)-1024:]), "%%EOF")
			},
		},
		{
			name: "should_render_a_month_with_no_expenses",
			fn: func(t *testing.T) {
				// The empty month is the one the scheduled send will hit first
				// if it ever runs against a fresh account.
				out, err := report.Render(logic.MonthlyReport{
					Month: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
				})
				require.NoError(t, err)
				require.Equal(t, "%PDF", string(out[:4]))
			},
		},
		{
			name: "should_render_an_ungrouped_report",
			fn: func(t *testing.T) {
				data := sampleReport()
				data.Grouped = false
				data.Sections = []logic.ReportSection{{
					Total:    36000,
					Expenses: data.Sections[0].Expenses,
				}}

				out, err := report.Render(data)
				require.NoError(t, err)
				require.Equal(t, "%PDF", string(out[:4]))
			},
		},
		{
			name: "should_render_accented_and_over_long_text",
			fn: func(t *testing.T) {
				// cp1252 covers the accents; the long description exercises the
				// truncation that keeps a cell inside its column.
				data := sampleReport()
				data.Sections[0].Expenses = []logic.ReportExpense{{
					Description:  "Almuerzo en la cafetería del aeropuerto international con varios acompañantes",
					CategoryName: "Alimentación",
					Amount:       30000,
				}}

				out, err := report.Render(data)
				require.NoError(t, err)
				require.Equal(t, "%PDF", string(out[:4]))
			},
		},
		{
			name: "should_paginate_a_long_month",
			fn: func(t *testing.T) {
				data := sampleReport()

				expenses := make([]logic.ReportExpense, 0, 400)
				for i := range 400 {
					expenses = append(expenses, logic.ReportExpense{
						Description:  "Expense",
						CategoryName: "Food",
						Amount:       uint64(i) + 1,
					})
				}
				data.Sections[0].Expenses = expenses

				out, err := report.Render(data)
				require.NoError(t, err)

				matches := pageCountRE.FindSubmatch(out)
				require.NotNil(t, matches, "no page count in the document")

				pages, err := strconv.Atoi(string(matches[1]))
				require.NoError(t, err)
				require.GreaterOrEqual(t, pages, 4,
					"400 rows should not fit on fewer than four pages")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

func TestTitle(t *testing.T) {
	require.Equal(t, "Expense report — September 2026", report.Title(sampleReport()))
}
