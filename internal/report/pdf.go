// Package report renders the monthly expense report as a PDF.
//
// It is deliberately pure: it takes an assembled logic.MonthlyReport and
// returns bytes, touching neither HTTP nor the database, so the layout can be
// exercised against a literal and the arithmetic against the database
// separately (docs/monthly-report.md, Phase 2).
//
// The category chart is drawn as filled rectangles rather than rendered by a
// charting library. Chart.js, which draws the same figure in the SPA, is a
// browser library; the alternative — driving headless Chrome over an HTML
// page — would add a large external binary and fight the systemd unit's
// ProtectSystem=strict sandbox for a bar chart.
package report

import (
	"bytes"
	"fmt"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/go-pdf/fpdf"
)

// Page geometry, in millimetres on A4 portrait.
const (
	pageMargin  = 15.0
	contentWide = 210.0 - 2*pageMargin
	lineHeight  = 5.5
	rowHeight   = 6.0
)

// Column widths for the expense table, left to right, summing to contentWide.
const (
	descriptionWidth = 95.0
	categoryWidth    = 45.0
	amountWidth      = contentWide - descriptionWidth - categoryWidth
)

// Column widths for the budget table.
const (
	budgetNameWidth = 60.0
	budgetPctWidth  = 20.0
	budgetCellWidth = (contentWide - budgetNameWidth - budgetPctWidth) / 3
)

// Geometry of the category chart: a label, the bar filling what is left, and
// the amount right-aligned after it.
const (
	categoryLabelWidth = 45.0
	categoryValueWidth = 30.0
	categoryBarWidth   = contentWide - categoryLabelWidth - categoryValueWidth
	categoryBarHeight  = 4.0
)

// The palette, chosen to survive a monochrome printer: one mid-blue for the
// bars, near-black for text, grey for secondary text and rules, and a red used
// only to mark a budget that has been exceeded.
//
//nolint:gochecknoglobals // immutable palette; an array cannot be a Go constant
var (
	inkColor    = [3]int{28, 28, 32}
	mutedColor  = [3]int{110, 110, 120}
	ruleColor   = [3]int{205, 205, 212}
	barColor    = [3]int{45, 110, 176}
	barBedColor = [3]int{228, 232, 238}
	overColor   = [3]int{176, 48, 60}
)

// Render draws the report and returns the PDF bytes.
func Render(report logic.MonthlyReport) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pageMargin, pageMargin, pageMargin)
	pdf.SetAutoPageBreak(true, pageMargin)
	pdf.SetTitle(Title(report), true)
	pdf.AddPage()

	// The core fonts are Latin-1, so a description typed with an accent would
	// otherwise reach the page as mojibake. cp1252 covers the accented Latin
	// letters an expense description realistically carries; anything outside
	// it is dropped by the translator rather than corrupting the stream.
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	d := &drawer{pdf: pdf, tr: tr}

	d.header(report)
	d.summary(report)
	d.categoryChart(report.CategoryTotals)
	d.sections(report)
	d.budgets(report.Budgets)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderFailed, err)
	}

	return buf.Bytes(), nil
}

// Title is the report's human name, used for the PDF metadata and by the
// handler for the download filename.
func Title(report logic.MonthlyReport) string {
	return "Expense report — " + report.Month.Format("January 2006")
}

// drawer holds the cursor-style state fpdf works with, so the section
// functions read as a sequence of blocks rather than as coordinate arithmetic.
type drawer struct {
	pdf *fpdf.Fpdf
	tr  func(string) string
}

func (d *drawer) setColor(c [3]int) {
	d.pdf.SetTextColor(c[0], c[1], c[2])
}

func (d *drawer) setFill(c [3]int) {
	d.pdf.SetFillColor(c[0], c[1], c[2])
}

func (d *drawer) text(s string, width, height float64, align string) {
	d.pdf.CellFormat(width, height, d.tr(s), "", 0, align, false, 0, "")
}

func (d *drawer) line(s string, height float64) {
	d.pdf.CellFormat(contentWide, height, d.tr(s), "", 1, "L", false, 0, "")
}

// rule draws the hairline that separates blocks, and doubles as the vertical
// spacing between them.
func (d *drawer) rule() {
	d.pdf.Ln(2)
	d.pdf.SetDrawColor(ruleColor[0], ruleColor[1], ruleColor[2])
	y := d.pdf.GetY()
	d.pdf.Line(pageMargin, y, pageMargin+contentWide, y)
	d.pdf.Ln(4)
}

// heading starts a block. keepWith reserves the space the block's first rows
// need, so a heading cannot be orphaned at the foot of a page.
func (d *drawer) heading(s string, keepWith float64) {
	d.keepTogether(lineHeight + keepWith)
	d.pdf.SetFont("Helvetica", "B", 11)
	d.setColor(inkColor)
	d.line(s, lineHeight)
	d.pdf.Ln(1)
}

// keepTogether starts a new page when less than height remains, so a group of
// rows that belongs together is not split across the break.
func (d *drawer) keepTogether(height float64) {
	_, pageHeight := d.pdf.GetPageSize()
	_, _, _, bottom := d.pdf.GetMargins()

	if d.pdf.GetY()+height > pageHeight-bottom {
		d.pdf.AddPage()
	}
}

func (d *drawer) header(report logic.MonthlyReport) {
	d.pdf.SetFont("Helvetica", "B", 20)
	d.setColor(inkColor)
	d.line("Expense report", 9)

	d.pdf.SetFont("Helvetica", "", 12)
	d.setColor(mutedColor)
	d.line(report.Month.Format("January 2006"), lineHeight)
	d.rule()
}

func (d *drawer) summary(report logic.MonthlyReport) {
	d.pdf.SetFont("Helvetica", "B", 16)
	d.setColor(inkColor)
	d.line(FormatCents(report.Total), 8)

	d.pdf.SetFont("Helvetica", "", 9)
	d.setColor(mutedColor)
	d.line(pluralize(report.ExpenseCount, "expense", "expenses")+" billed to this month", lineHeight)
	d.rule()
}

func (d *drawer) categoryChart(totals []logic.ReportCategoryTotal) {
	if len(totals) == 0 {
		return
	}

	d.heading("Spending by category", rowHeight*2)

	// The bars are scaled against the largest category rather than against the
	// month's total: with one dominant category every other bar would round to
	// nothing.
	largest := totals[0].Total
	for _, total := range totals {
		if total.Total > largest {
			largest = total.Total
		}
	}

	d.pdf.SetFont("Helvetica", "", 9)

	for _, total := range totals {
		d.keepTogether(rowHeight)

		y := d.pdf.GetY()

		d.setColor(inkColor)
		d.text(truncate(d.pdf, total.Name, categoryLabelWidth-2), categoryLabelWidth, rowHeight, "L")

		barTop := y + (rowHeight-categoryBarHeight)/2
		d.setFill(barBedColor)
		d.pdf.Rect(pageMargin+categoryLabelWidth, barTop, categoryBarWidth, categoryBarHeight, "F")

		if largest > 0 {
			width := categoryBarWidth * float64(total.Total) / float64(largest)
			d.setFill(barColor)
			d.pdf.Rect(pageMargin+categoryLabelWidth, barTop, width, categoryBarHeight, "F")
		}

		d.pdf.SetXY(pageMargin+categoryLabelWidth+categoryBarWidth, y)
		d.setColor(inkColor)
		d.pdf.CellFormat(categoryValueWidth, rowHeight, d.tr(FormatCents(total.Total)), "", 1, "R", false, 0, "")
	}

	d.rule()
}

func (d *drawer) sections(report logic.MonthlyReport) {
	if len(report.Sections) == 0 {
		d.pdf.SetFont("Helvetica", "", 10)
		d.setColor(mutedColor)
		d.line("No expenses billed to this month.", lineHeight)

		return
	}

	d.heading("Expenses", rowHeight*3)
	// Printed once rather than per section: the columns do not change, and a
	// header under every tag turned the list into stripes.
	d.expenseTableHead()

	for _, section := range report.Sections {
		// An ungrouped report has one nameless section, so its header row is
		// skipped rather than printed blank.
		if section.Name != "" {
			d.keepTogether(rowHeight * 3)
			d.pdf.SetFont("Helvetica", "B", 10)
			d.setColor(inkColor)
			nameWidth := descriptionWidth + categoryWidth
			d.text(truncate(d.pdf, section.Name, nameWidth-2), nameWidth, rowHeight, "L")
			d.pdf.CellFormat(amountWidth, rowHeight, d.tr(FormatCents(section.Total)), "", 1, "R", false, 0, "")
		}

		for _, expense := range section.Expenses {
			d.keepTogether(rowHeight)
			d.pdf.SetFont("Helvetica", "", 9)
			d.setColor(inkColor)
			d.text(truncate(d.pdf, expense.Description, descriptionWidth-2), descriptionWidth, rowHeight, "L")
			d.setColor(mutedColor)
			d.text(truncate(d.pdf, expense.CategoryName, categoryWidth-2), categoryWidth, rowHeight, "L")
			d.setColor(inkColor)
			d.pdf.CellFormat(amountWidth, rowHeight, d.tr(FormatCents(expense.Amount)), "", 1, "R", false, 0, "")
		}

		d.pdf.Ln(3)
	}

	d.pdf.SetFont("Helvetica", "B", 11)
	d.setColor(inkColor)
	d.text("Total", descriptionWidth+categoryWidth, rowHeight, "L")
	d.pdf.CellFormat(amountWidth, rowHeight, d.tr(FormatCents(report.Total)), "", 1, "R", false, 0, "")
	d.rule()
}

func (d *drawer) expenseTableHead() {
	d.pdf.SetFont("Helvetica", "", 8)
	d.setColor(mutedColor)
	d.text("Description", descriptionWidth, lineHeight, "L")
	d.text("Category", categoryWidth, lineHeight, "L")
	d.pdf.CellFormat(amountWidth, lineHeight, d.tr("Amount"), "", 1, "R", false, 0, "")
}

func (d *drawer) budgets(budgets []logic.ReportBudget) {
	if len(budgets) == 0 {
		return
	}

	d.heading("Budgets", rowHeight*2)

	d.pdf.SetFont("Helvetica", "", 8)
	d.setColor(mutedColor)
	d.text("Category", budgetNameWidth, lineHeight, "L")
	d.text("Spent", budgetCellWidth, lineHeight, "R")
	d.text("Budget", budgetCellWidth, lineHeight, "R")
	d.text("Left", budgetCellWidth, lineHeight, "R")
	d.pdf.CellFormat(budgetPctWidth, lineHeight, d.tr("Used"), "", 1, "R", false, 0, "")

	for _, budget := range budgets {
		d.keepTogether(rowHeight)

		d.pdf.SetFont("Helvetica", "", 9)
		d.setColor(inkColor)
		d.text(truncate(d.pdf, budget.CategoryName, budgetNameWidth-2), budgetNameWidth, rowHeight, "L")
		d.text(FormatCents(budget.Total), budgetCellWidth, rowHeight, "R")
		d.setColor(mutedColor)
		d.text(FormatCents(budget.Budget), budgetCellWidth, rowHeight, "R")

		// Only the two columns that carry the verdict change colour, so a row
		// reads as over budget without the whole line turning red.
		if budget.Over {
			d.setColor(overColor)
		} else {
			d.setColor(inkColor)
		}

		d.text(FormatSignedCents(budget.Left), budgetCellWidth, rowHeight, "R")
		d.pdf.CellFormat(budgetPctWidth, rowHeight, d.tr(fmt.Sprintf("%d%%", budget.Pct)), "", 1, "R", false, 0, "")
	}
}

// truncate shortens s with an ellipsis until it fits width, so a long
// description pushes nothing out of its column. fpdf clips silently otherwise,
// which reads as a missing word rather than as a cut.
func truncate(pdf *fpdf.Fpdf, s string, width float64) string {
	if pdf.GetStringWidth(s) <= width {
		return s
	}

	runes := []rune(s)
	for len(runes) > 1 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + "..."

		if pdf.GetStringWidth(candidate) <= width {
			return candidate
		}
	}

	return string(runes)
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}

	return fmt.Sprintf("%d %s", n, plural)
}
