// Package repo provides database access and query execution utilities.
package repo

import "database/sql"

// Queries wraps a sql.DB connection and provides methods for executing database queries.
type Queries struct {
	db *sql.DB
}

// New creates a new instance of Queries using the provided sql.DB connection.
func New(db *sql.DB) Queries {
	return Queries{
		db: db,
	}
}
