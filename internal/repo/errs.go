package repo

import (
	"errors"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrInvalidConnector  = errors.New("invalid operator")
	ErrInvalidOperator   = errors.New("invalid operator")
	ErrInvalidField      = errors.New("invalid field")
	ErrInvalidPagination = errors.New("invalid pagination values")
	ErrUnknownTaggable   = errors.New("unknown taggable")

	// The two sort refusals are the only QueryOptions errors a client can
	// cause — every filter field is chosen by the handler, the sort comes from
	// the query string — so the API answers them as 422s quoting these
	// messages. Keep them free of table or column detail.
	ErrInvalidSortField = errors.New("sort_field is missing or names a column this list cannot be sorted by")
	ErrInvalidSortOrder = errors.New(`sort_order must be "ASC" or "DESC"`)
)

// IsUniqueViolation reports whether err is SQLite's UNIQUE constraint failure.
// Callers use it to answer with a domain error instead of letting the driver's
// message — which names the table and column that collided — reach a page.
func IsUniqueViolation(err error) bool {
	var sqliteErr sqlite3.Error

	return errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
}
