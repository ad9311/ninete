package repo

import (
	"fmt"
	"slices"
	"strings"
)

type FilterField struct {
	Name     string
	Value    any
	Operator string
}

type Filters struct {
	FilterFields []FilterField
	Connector    string
}

type Sorting struct {
	Field string
	Order string
}

type Pagination struct {
	PerPage int
	Page    int
}

type QueryOptions struct {
	Filters    Filters
	Sorting    Sorting
	Pagination Pagination
	filters    string
	sorting    string
	pagination string
}

func (q *QueryOptions) Build() (string, error) {
	filters, err := q.Filters.Build()
	if err != nil {
		return "", err
	}

	sorting, err := q.Sorting.Build()
	if err != nil {
		return "", err
	}

	pagination, err := q.Pagination.Build()
	if err != nil {
		return "", err
	}

	q.filters = filters
	q.sorting = sorting
	q.pagination = pagination

	subQuery := q.filters + " " + q.sorting + " " + q.pagination

	return strings.TrimSpace(subQuery), nil
}

func (q *QueryOptions) FilterSubQuery() string {
	return q.filters
}

func (q *QueryOptions) SortingSubQuery() string {
	return q.sorting
}

func (q *QueryOptions) PaginationSubQuery() string {
	return q.pagination
}

func (q *QueryOptions) Validate(fields []string) error {
	if !q.Filters.ValidFields(fields) || !q.Sorting.ValidField(fields) {
		availableFields := strings.Join(fields, ",")

		return fmt.Errorf("%w, valid fields for filters and sorting are: %s", ErrInvalidField, availableFields)
	}

	return nil
}

func (f *Filters) Build() (string, error) {
	if len(f.FilterFields) == 0 {
		return "", nil
	}

	if len(f.FilterFields) > 1 && !f.validConnector() {
		return "", fmt.Errorf("%w, valid connectors are 'AND' or 'OR'", ErrInvalidConnector)
	}

	var buildFilters []string

	for _, field := range f.FilterFields {
		if field.Name == "" || !field.validOperator() {
			operators := strings.Join(validOperators(), ", ")

			return "", fmt.Errorf("%w, valid operators are %s", ErrInvalidOperator, operators)
		}

		filter := fmt.Sprintf("\"%s\" %s %s", field.Name, field.Operator, "?")

		buildFilters = append(buildFilters, filter)
	}

	joined := strings.Join(buildFilters, " "+f.Connector+" ")

	return "WHERE " + joined, nil
}

func (f *Filters) validConnector() bool {
	f.Connector = strings.ToUpper(f.Connector)
	if f.Connector != "AND" && f.Connector != "OR" {
		return false
	}

	return true
}

func (f *Filters) ValidFields(fields []string) bool {
	if len(f.FilterFields) == 0 {
		return true
	}

	for _, field := range f.FilterFields {
		if !slices.Contains(fields, field.Name) {
			return false
		}
	}

	return true
}

func (f *FilterField) validOperator() bool {
	return slices.Contains(validOperators(), f.Operator)
}

func (f *Filters) Values() []any {
	var values []any

	for _, v := range f.FilterFields {
		values = append(values, v.Value)
	}

	return values
}

func validOperators() []string {
	return []string{"=", ">", "<", ">=", "<="}
}

func (s *Sorting) Build() (string, error) {
	if s.Field == "" && s.Order == "" {
		return "", nil
	}

	if !s.validateSortOrder() {
		return "", fmt.Errorf("%w, valid order are 'ASC' or 'DESC'", ErrInvalidSortOrder)
	}

	sorting := fmt.Sprintf("ORDER BY \"%s\" %s", s.Field, s.Order)

	return sorting, nil
}

func (s *Sorting) validateSortOrder() bool {
	s.Order = strings.ToUpper(s.Order)

	if s.Order != "ASC" && s.Order != "DESC" {
		return false
	}

	return true
}

func (s *Sorting) ValidField(fields []string) bool {
	if s.Field == "" && s.Order == "" {
		return true
	}

	return slices.Contains(fields, s.Field)
}

func (p *Pagination) Build() (string, error) {
	if p.PerPage == 0 && p.Page == 0 {
		return "", nil
	}

	if p.Page < 1 || p.PerPage < 1 {
		return "", ErrInvalidPagination
	}

	offset := (p.Page - 1) * p.PerPage
	paginate := fmt.Sprintf("LIMIT %d OFFSET %d", p.PerPage, offset)

	return paginate, nil
}
