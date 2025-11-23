package repo

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/prog"
)

type Queries struct {
	app *prog.App
	db  *sql.DB
}

type DBConnStats struct {
	MaxOpenConnections int `json:"maxOpenConnections"`
	IdleConnections    int `json:"idleConnections"`
	InUseConnections   int `json:"inUseConnections"`
}

type FilterField struct {
	Name     string `json:"name"`
	Value    any    `json:"value"`
	Operator string `json:"operator"`
}

type Filters struct {
	FilterFields []FilterField `json:"fields"`
	Connector    string        `json:"connector"`
}

type Sorting struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

type Pagination struct {
	PerPage int `json:"perPage"`
	Page    int `json:"page"`
}

type QueryOptions struct {
	Filters    Filters    `json:"filters"`
	Sorting    Sorting    `json:"sorting"`
	Pagination Pagination `json:"pagination"`
	filters    string
	sorting    string
	pagination string
}

func New(app *prog.App, db *sql.DB) Queries {
	return Queries{
		app: app,
		db:  db,
	}
}

func (q *Queries) CheckDBStatus() (DBConnStats, error) {
	var stats DBConnStats

	if err := q.db.Ping(); err != nil {
		return stats, fmt.Errorf("failed to ping database: %w", err)
	}

	stats = DBConnStats{
		MaxOpenConnections: q.db.Stats().MaxOpenConnections,
		IdleConnections:    q.db.Stats().Idle,
		InUseConnections:   q.db.Stats().InUse,
	}

	return stats, nil
}

func (q *Queries) wrapQuery(query string, queryFunc func() error) error {
	if !q.app.Logger.EnableQuery {
		err := queryFunc()

		return err
	}

	start := time.Now()
	defer func() {
		q.app.Logger.Query(query, time.Since(start))
	}()

	return queryFunc()
}

func newUpdatedAt() int64 {
	return time.Now().UTC().Unix()
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

	return subQuery, nil
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

func (f *Filters) Build() (string, error) {
	if len(f.FilterFields) == 0 {
		return "", nil
	}

	if len(f.FilterFields) > 1 && !f.validConnector() {
		return "", ErrInvalidConnector
	}

	var buildFilters []string

	for _, field := range f.FilterFields {
		if field.Name == "" || !field.validOperator() {
			return "", ErrInvalidOperator
		}

		filter := fmt.Sprintf("\"%s\" %s %s", field.Name, field.Operator, "?")

		buildFilters = append(buildFilters, filter)
	}

	joined := strings.Join(buildFilters, " "+f.Connector+" ")

	return "WHERE " + joined, nil
}

func (f *Filters) validConnector() bool {
	if f.Connector != "AND" && f.Connector != "OR" {
		return false
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
		return "", ErrInvalidSortOrder
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
