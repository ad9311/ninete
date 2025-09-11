// Package db provides utilities for opening and managing database connections.
package db

import (
	"database/sql"
	"fmt"
	"log"
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

// Pool wraps a standard sql.DB connection pool, providing a convenient way to manage and share
// database connections throughout the application.
type Pool struct {
	DB *sql.DB
}

// dbConf holds the configuration parameters for the database connection.
type dbConf struct {
	URL          string
	MaxOpenConns int
	MaxIdleConns int
}

// Open initializes and returns a new database connection using the provided configuration.
// It connects to a SQLite3 database specified by the URL in the conf.dbConf struct.
func Open(_ string) (*Pool, error) {
	var conn *Pool

	dc, err := loadConf()
	if err != nil {
		return conn, err
	}

	sqlDB, err := sql.Open("sqlite3", "file:"+dc.URL)
	if err != nil {
		return conn, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return conn, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDB.SetMaxOpenConns(dc.MaxOpenConns)
	sqlDB.SetMaxIdleConns(dc.MaxIdleConns)

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

// loadDBConf loads the database configuration from environment variables.
func loadConf() (dbConf, error) {
	var dbc dbConf

	envName := "DATABASE_URL"
	url := os.Getenv(envName)
	if url == "" {
		return dbc, fmt.Errorf("%w: %s", errs.ErrEnvNoTSet, envName)
	}

	maxOpenConns, err := setInt("MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return dbc, err
	}

	maxIdleConns, err := setInt("MAX_IDLE_CONNS", defaultMaxIdleConns)
	if err != nil {
		return dbc, err
	}

	dbc = dbConf{
		URL:          url,
		MaxOpenConns: maxOpenConns,
		MaxIdleConns: maxIdleConns,
	}

	return dbc, nil
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
