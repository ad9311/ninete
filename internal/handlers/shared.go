package handlers

import (
	"net/http"
	"strconv"

	"github.com/ad9311/ninete/internal/repo"
)

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
	Done        string
	Priority    int
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
	if dr, ok := computeDateRange(dateRangeKey); ok {
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
	priority, _ := strconv.Atoi(q.Get("priority"))

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
		Done:        q.Get("done"),
		Priority:    priority,
	}
}

func tagNamesByTargetID(rows []repo.TagRow) map[int][]string {
	m := map[int][]string{}
	for _, row := range rows {
		m[row.TargetID] = append(m[row.TargetID], row.TagName)
	}

	return m
}
