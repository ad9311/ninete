// Package db provides utilities for opening and managing database connections.
package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/ad9311/ninete/internal/prog"
	_ "github.com/mattn/go-sqlite3" // Database driver
)

// Database connection pool configuration constants.
const (
	defaultMaxOpenConns = 1
	defaultMaxIdleConns = 1
)

// Open initializes and returns a new database connection using the provided configuration.
// It connects to a SQLite3 database specified by the url in the conf.conf struct.
func Open() (*sql.DB, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, fmt.Errorf("%w: DATABASE_URL", prog.ErrEnvNoTSet)
	}

	maxOpenConns, err := prog.SetInt("MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return nil, err
	}

	maxIdleConns, err := prog.SetInt("MAX_IDLE_CONNS", defaultMaxIdleConns)
	if err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open("sqlite3", "file:"+url)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)

	return sqlDB, nil
}
