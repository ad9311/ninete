package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

type expenseFormBase struct {
	CategoryID  int
	Description string
	Amount      uint64
}

type dateRange struct {
	start int64
	end   int64
}

func parseTZOffset(r *http.Request) int {
	offset, _ := strconv.Atoi(r.URL.Query().Get("tz_offset"))

	return offset
}

var dateRangeLabels = []struct { //nolint:gochecknoglobals // static lookup table
	Value string
	Label string
}{
	{"this_month", "This month"},
	{"next_month", "Next month"},
	{"last_month", "Last month"},
	{"this_week", "This week"},
	{"six_months", "Last 6 months"},
	{"this_year", "This year"},
}

func DateRangeOptions() []struct {
	Value string
	Label string
} {
	return dateRangeLabels
}

// budgetMode names how /expenses/budgets renders a range: one bar per category
// for a single month, or a per-month breakdown for a multi-month span.
type budgetMode string

const (
	budgetModeMonth  budgetMode = "month"
	budgetModeMonths budgetMode = "months"
)

// budgetDateRanges is the subset of computeDateRange's keys the budgets page
// offers. this_week is excluded because seven days read against a monthly
// budget is a false comparison, next_month because it has no spending yet, and
// all_time because the month count is unbounded.
var budgetDateRanges = []struct { //nolint:gochecknoglobals // static lookup table
	Value string
	Label string
	Mode  budgetMode
}{
	{"this_month", "This month", budgetModeMonth},
	{"last_month", "Last month", budgetModeMonth},
	{"six_months", "Last 6 months", budgetModeMonths},
	{"this_year", "This year", budgetModeMonths},
}

// budgetDateRange normalizes a requested range key onto the supported set,
// falling back to this_month for anything else.
func budgetDateRange(key string) (string, budgetMode) {
	for _, r := range budgetDateRanges {
		if r.Value == key {
			return r.Value, r.Mode
		}
	}

	return budgetDateRanges[0].Value, budgetDateRanges[0].Mode
}

func computeDateRange(key string, tzOffsetMinutes int) (dateRange, bool) {
	loc := time.FixedZone("client", -tzOffsetMinutes*60)
	now := time.Now().In(loc)
	year, month, _ := now.Date()

	switch key {
	case "this_month":
		start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)

		return dateRange{start.Unix(), end.Unix()}, true
	case "next_month":
		start := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)

		return dateRange{start.Unix(), end.Unix()}, true
	case "last_month":
		start := time.Date(year, month-1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

		return dateRange{start.Unix(), end.Unix()}, true
	case "this_week":
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -int(weekday-time.Monday))
		start := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 7)

		return dateRange{start.Unix(), end.Unix()}, true
	case "six_months":
		// Five months back plus the current one is six, matching the label.
		// month-6 would span seven, which the budgets page prints as the
		// denominator of "N of M months over".
		start := time.Date(year, month-5, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)

		return dateRange{start.Unix(), end.Unix()}, true
	case "this_year":
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

		return dateRange{start.Unix(), end.Unix()}, true
	default:
		return dateRange{}, false
	}
}

// parseAPIDateBounds reads the explicit [start, end) bounds an /api/expenses*
// request carries instead of date_range+tz_offset (§3.6 of
// docs/spa-migration.md, "Retiring tz_offset on the API side"): the client
// already knows its own zone, so it resolves the named range to UTC-midnight
// epoch bounds before the fetch. Neither param present means "no bound" (the
// all_time case); either present alone, or start on or after end, is
// malformed input rather than a silent fallback.
func parseAPIDateBounds(q url.Values) (start, end int64, hasBounds bool, err error) {
	rawStart, rawEnd := q.Get("start"), q.Get("end")
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

func parseExpenseFormBase(r *http.Request) (expenseFormBase, error) {
	var base expenseFormBase

	if err := r.ParseForm(); err != nil {
		return base, fmt.Errorf("%w: %w", ErrParseForm, err)
	}

	categoryID, err := prog.ParseID(r.FormValue("category_id"), "Category ID")
	if err != nil {
		return base, err
	}

	amount, err := prog.ParseAmount(r.FormValue("amount"))
	if err != nil {
		return base, err
	}

	base.CategoryID = categoryID
	base.Description = r.FormValue("description")
	base.Amount = amount

	return base, nil
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

func (h *Handler) findCategoriesOrErr(
	w http.ResponseWriter,
	r *http.Request,
	tmpl TemplateName,
) ([]repo.Category, map[int]string, bool) {
	categories, nameByID, err := h.findCategories(r.Context())
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, tmpl, err)

		return nil, nil, false
	}

	return categories, nameByID, true
}

func categoryNameOrUnknown(nameByID map[int]string, categoryID int) string {
	if name := nameByID[categoryID]; name != "" {
		return name
	}

	return "Unknown"
}

func setResourceFormData(
	data map[string]any,
	categories []repo.Category,
	resourceName string,
	resource any,
) {
	data["categories"] = categories
	data[resourceName] = resource
}
