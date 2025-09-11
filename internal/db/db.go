// Package db provides utilities for opening and managing database connections.
package db

import (
	"database/sql"

	"github.com/ad9311/ninete/internal/conf"
	_ "github.com/mattn/go-sqlite3" // Database driver
)

// Open initializes and returns a new database connection using the provided configuration.
// It connects to a SQLite3 database specified by the URL in the conf.DBConf struct.
func Open(dbc conf.DBConf) (*sql.DB, error) {
	var sqlDB *sql.DB

	sqlDB, err := sql.Open("sqlite3", "file:"+dbc.URL)
	if err != nil {
		return sqlDB, err
	}

	sqlDB.SetMaxOpenConns(dbc.MaxOpenConns)
	sqlDB.SetMaxIdleConns(dbc.MaxIdleConns)

	return sqlDB, nil
}
