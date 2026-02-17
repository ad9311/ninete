package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

const defaultPerPage = 10

type expenseFormBase struct {
	CategoryID  int
	Description string
	Amount      uint64
}

type PaginationData struct {
	CurrentPage int
	TotalPages  int
	PerPage     int
	TotalCount  int
	HasPrev     bool
	HasNext     bool
	SortField   string
	SortOrder   string
	CategoryID  int
	DateRange   string
}

type dateRange struct {
	start int64
	end   int64
}

var dateRangeLabels = []struct { //nolint:gochecknoglobals // static lookup table
	Value string
	Label string
}{
	{"this_month", "This month"},
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

func computeDateRange(key string) (dateRange, bool) {
	now := time.Now()
	year, month, _ := now.Date()
	loc := now.Location()

	switch key {
	case "this_month":
		start := time.Date(year, month, 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 1, 0)

		return dateRange{start.Unix(), end.Unix()}, true
	case "last_month":
		start := time.Date(year, month-1, 1, 0, 0, 0, 0, loc)
		end := time.Date(year, month, 1, 0, 0, 0, 0, loc)

		return dateRange{start.Unix(), end.Unix()}, true
	case "this_week":
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -int(weekday-time.Monday))
		start := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 7)

		return dateRange{start.Unix(), end.Unix()}, true
	case "six_months":
		start := time.Date(year, month-6, 1, 0, 0, 0, 0, loc)
		end := time.Date(year, month, 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0)

		return dateRange{start.Unix(), end.Unix()}, true
	case "this_year":
		start := time.Date(year, 1, 1, 0, 0, 0, 0, loc)
		end := time.Date(year+1, 1, 1, 0, 0, 0, 0, loc)

		return dateRange{start.Unix(), end.Unix()}, true
	default:
		return dateRange{}, false
	}
}

func userScopedQueryOpts(r *http.Request, userID int, defaultSort repo.Sorting) repo.QueryOptions {
	q := r.URL.Query()

	sorting := repo.Sorting{
		Order: q.Get("sort_order"),
		Field: q.Get("sort_field"),
	}
	if sorting.Field == "" && sorting.Order == "" {
		sorting = defaultSort
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 {
		perPage = defaultPerPage
	}

	opts := repo.QueryOptions{
		Sorting: sorting,
		Pagination: repo.Pagination{
			Page:    page,
			PerPage: perPage,
		},
	}
	opts.Filters.FilterFields = append(opts.Filters.FilterFields, repo.FilterField{
		Name:     "user_id",
		Value:    userID,
		Operator: "=",
	})
	opts.Filters.Connector = "AND"

	if categoryID, _ := strconv.Atoi(q.Get("category_id")); categoryID > 0 {
		opts.Filters.FilterFields = append(opts.Filters.FilterFields, repo.FilterField{
			Name:     "category_id",
			Value:    categoryID,
			Operator: "=",
		})
	}

	if dr, ok := computeDateRange(q.Get("date_range")); ok {
		opts.Filters.FilterFields = append(opts.Filters.FilterFields,
			repo.FilterField{Name: "date", Value: dr.start, Operator: ">="},
			repo.FilterField{Name: "date", Value: dr.end, Operator: "<"},
		)
	}

	return opts
}

func newPaginationData(r *http.Request, opts repo.QueryOptions, totalCount int) PaginationData {
	totalPages := 0
	if opts.Pagination.PerPage > 0 {
		totalPages = (totalCount + opts.Pagination.PerPage - 1) / opts.Pagination.PerPage
	}

	q := r.URL.Query()
	categoryID, _ := strconv.Atoi(q.Get("category_id"))

	return PaginationData{
		CurrentPage: opts.Pagination.Page,
		TotalPages:  totalPages,
		PerPage:     opts.Pagination.PerPage,
		TotalCount:  totalCount,
		HasPrev:     opts.Pagination.Page > 1,
		HasNext:     opts.Pagination.Page < totalPages,
		SortField:   opts.Sorting.Field,
		SortOrder:   opts.Sorting.Order,
		CategoryID:  categoryID,
		DateRange:   q.Get("date_range"),
	}
}

func parseExpenseFormBase(r *http.Request) (expenseFormBase, error) {
	var base expenseFormBase

	if err := r.ParseForm(); err != nil {
		return base, fmt.Errorf("failed to parse form, %w", err)
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
