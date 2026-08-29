package repo

import (
	"errors"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrInvalidConnector  = errors.New("invalid operator")
	ErrInvalidOperator   = errors.New("invalid operator")
	ErrInvalidField      = errors.New("invalid field")
	ErrInvalidSortOrder  = errors.New("invalid sort order")
	ErrInvalidPagination = errors.New("invalid pagination values")
	ErrUnknownTaggable   = errors.New("unknown taggable")
)

// IsUniqueViolation reports whether err is SQLite's UNIQUE constraint failure.
// Callers use it to answer with a domain error instead of letting the driver's
// message — which names the table and column that collided — reach a page.
func IsUniqueViolation(err error) bool {
	var sqliteErr sqlite3.Error

	return errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
}
