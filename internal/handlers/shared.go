package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ad9311/ninete/internal/repo"
)

func parseFloatField(r *http.Request, field string) (float64, error) {
	v, err := strconv.ParseFloat(r.FormValue(field), 64)
	if err != nil {
		return 0, fmt.Errorf("%w %q: %w", ErrParseField, field, err)
	}

	return v, nil
}

const defaultPerPage = 10

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

func userScopedQueryOpts(
	r *http.Request, userID int, defaultSort repo.Sorting, defaultDateRange string,
) repo.QueryOptions {
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

	dateRangeKey := q.Get("date_range")
	if dateRangeKey == "" {
		dateRangeKey = defaultDateRange
	}
	if dr, ok := computeDateRange(dateRangeKey, parseTZOffset(r)); ok {
		opts.Filters.FilterFields = append(opts.Filters.FilterFields,
			repo.FilterField{Name: "date", Value: dr.start, Operator: ">="},
			repo.FilterField{Name: "date", Value: dr.end, Operator: "<"},
		)
	}

	return opts
}

func newPaginationData(r *http.Request, opts repo.QueryOptions, totalCount int, defaultDateRange string) PaginationData { //nolint:lll
	totalPages := 0
	if opts.Pagination.PerPage > 0 {
		totalPages = (totalCount + opts.Pagination.PerPage - 1) / opts.Pagination.PerPage
	}

	q := r.URL.Query()
	categoryID, _ := strconv.Atoi(q.Get("category_id"))

	dateRange := q.Get("date_range")
	if dateRange == "" {
		dateRange = defaultDateRange
	}

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
		DateRange:   dateRange,
	}
}

func safeUint64ToInt(v uint64) int {
	const maxInt = uint64(^uint(0) >> 1)
	if v > maxInt {
		return int(maxInt)
	}

	return int(v)
}

func nextSortOrder(currentField, currentOrder, column, columnDefault string) string {
	if currentField != column {
		return columnDefault
	}

	if currentOrder == "ASC" {
		return "DESC"
	}

	return "ASC"
}
