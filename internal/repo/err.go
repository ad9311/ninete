package repo

import "errors"

var (
	ErrInvalidConnector  = errors.New("invalid operator")
	ErrInvalidOperator   = errors.New("invalid operator")
	ErrInvalidField      = errors.New("invalid field")
	ErrInvalidSortField  = errors.New("invalid sort field")
	ErrInvalidSortOrder  = errors.New("invalid sort order")
	ErrInvalidPagination = errors.New("invalid pagination values")
)
