package handlers

import (
	"slices"
	"strconv"

	"github.com/ad9311/ninete/internal/repo"
)

const defaultPerPage = 15

// perPageChoices are the only page sizes the listings accept. Anything else in
// the query string falls back to defaultPerPage, so a hand-edited per_page
// cannot ask the database for an unbounded page.
var perPageChoices = []int{15, 25, 50, 100} //nolint:gochecknoglobals // static option list

func normalizePerPage(raw string) int {
	perPage, err := strconv.Atoi(raw)
	if err != nil || !slices.Contains(perPageChoices, perPage) {
		return defaultPerPage
	}

	return perPage
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
	Search      string
	Tag         string
	DateFrom    string
	DateTo      string
}

func userScopedQueryOpts(q apiListQuery, userID int, defaultSort repo.Sorting) repo.QueryOptions {
	sorting := repo.Sorting{
		Order: q.SortOrder,
		Field: q.SortField,
	}
	if sorting.Field == "" && sorting.Order == "" {
		sorting = defaultSort
	}

	page, _ := strconv.Atoi(q.Page)
	if page < 1 {
		page = 1
	}

	perPage := normalizePerPage(q.PerPage)

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

	if categoryID, _ := strconv.Atoi(q.CategoryID); categoryID > 0 {
		opts.Filters.FilterFields = append(opts.Filters.FilterFields, repo.FilterField{
			Name:     "category_id",
			Value:    categoryID,
			Operator: "=",
		})
	}

	return opts
}

func newPaginationData(q apiListQuery, opts repo.QueryOptions, totalCount int) PaginationData {
	totalPages := 0
	if opts.Pagination.PerPage > 0 {
		totalPages = (totalCount + opts.Pagination.PerPage - 1) / opts.Pagination.PerPage
	}

	categoryID, _ := strconv.Atoi(q.CategoryID)

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
