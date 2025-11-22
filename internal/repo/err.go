package repo

import "errors"

var (
	ErrInvalidConnector = errors.New("invalid operator")
	ErrInvalidOperator  = errors.New("invalid operator")
	ErrEmptyField       = errors.New("empty name for field")
	ErrInvalidSortOrder = errors.New("invalid sort order")
)
