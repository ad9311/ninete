// Package db provides utilities for opening and managing database connections.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/ad9311/ninete/internal/errs"
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
		return nil, fmt.Errorf("%w: DATABASE_URL", errs.ErrEnvNoTSet)
	}

	maxOpenConns, err := setInt("MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return nil, err
	}

	maxIdleConns, err := setInt("MAX_IDLE_CONNS", defaultMaxIdleConns)
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

// setInt retrieves an integer value from the environment variable specified by envName.
// If the environment variable is not set, it returns the provided default value def.
func setInt(envName string, def int) (int, error) {
	maxConnsStr := os.Getenv(envName)
	if maxConnsStr == "" {
		return def, nil
	}
	v, err := strconv.ParseInt(maxConnsStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %w", maxConnsStr, err)
	}

	return int(v), nil
}
