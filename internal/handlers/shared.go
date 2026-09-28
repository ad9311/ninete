package handlers

import (
	"strconv"

	"github.com/ad9311/ninete/internal/repo"
)

const defaultPerPage = 15

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

// userScopedQueryOpts builds a listing's options from a query decodeQuery has
// already validated, so every value here is either absent or well formed: the
// page sizes per_page accepts are the oneof on apiListQuery, which is what
// keeps a hand-edited URL from asking the database for an unbounded page.
// An absent half of the sort pair takes the default's half, so a field sent
// alone sorts in the default direction rather than failing.
func userScopedQueryOpts(
	q apiListQuery, sortField, sortOrder string, userID int, defaultSort repo.Sorting,
) repo.QueryOptions {
	sorting := repo.Sorting{Field: sortField, Order: sortOrder}
	if sorting.Field == "" {
		sorting.Field = defaultSort.Field
	}
	if sorting.Order == "" {
		sorting.Order = defaultSort.Order
	}

	page := 1
	if q.Page != "" {
		page, _ = strconv.Atoi(q.Page)
	}

	perPage := defaultPerPage
	if q.PerPage != "" {
		perPage, _ = strconv.Atoi(q.PerPage)
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
