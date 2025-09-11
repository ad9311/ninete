// Package db provides utilities for opening and managing database connections.
package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ad9311/ninete/internal/conf"
	_ "github.com/mattn/go-sqlite3" // Database driver
)

// Pool wraps a standard sql.DB connection pool, providing a convenient way to manage and share
// database connections throughout the application.
type Pool struct {
	DB *sql.DB
}

// Open initializes and returns a new database connection using the provided configuration.
// It connects to a SQLite3 database specified by the URL in the conf.DBConf struct.
func Open(dbc conf.DBConf) (*Pool, error) {
	var conn *Pool

	sqlDB, err := sql.Open("sqlite3", "file:"+dbc.URL)
	if err != nil {
		return conn, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return conn, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDB.SetMaxOpenConns(dbc.MaxOpenConns)
	sqlDB.SetMaxIdleConns(dbc.MaxIdleConns)

	conn = &Pool{
		DB: sqlDB,
	}

	return conn, nil
}

// Close gracefully closes the database connection pool associated with the Pool instance.
// If an error occurs during the closing process, it is logged using the standard logger.
func (p *Pool) Close() {
	if err := p.DB.Close(); err != nil {
		log.Println(err)
	}
}
