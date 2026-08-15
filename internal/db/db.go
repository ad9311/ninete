package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"embed"
	"fmt"
	"os"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/mattn/go-sqlite3"
)

const (
	DefaultMaxOpenConns = 1
	DefaultMaxIdleConns = 1
)

const (
	databaseInitFile   = "init/database.sql"
	connectionInitFile = "init/connection.sql"
)

// driverName registers go-sqlite3 under our own name so the connect hook below
// runs for every connection the pool opens.
const driverName = "sqlite3_ninete"

//go:embed init/*.sql
var initPragmas embed.FS

func init() {
	sql.Register(driverName, &sqlite3.SQLiteDriver{ConnectHook: applyConnectionPragmas})
}

func Open() (*sql.DB, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, fmt.Errorf("'DATABASE_URL' %w", prog.ErrEnvNoTSet)
	}

	maxOpenConns, err := prog.SetInt("MAX_OPEN_CONNS", DefaultMaxOpenConns)
	if err != nil {
		return nil, err
	}

	maxIdleConns, err := prog.SetInt("MAX_IDLE_CONNS", DefaultMaxIdleConns)
	if err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open(driverName, "file:"+url+"?_loc=UTC")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	databaseQuery, err := readInitSQL(databaseInitFile)
	if err != nil {
		return nil, fmt.Errorf("failed to run %s script: %w", databaseInitFile, err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err := sqlDB.Exec(databaseQuery); err != nil {
		return nil, fmt.Errorf("failed to run database PRAGMA commands: %w", err)
	}

	if err := verifyEnvStamp(sqlDB); err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)

	return sqlDB, nil
}

// Optimize refreshes the query planner statistics. It belongs at shutdown: run
// at startup, as it used to be, PRAGMA optimize has no accumulated query history
// to learn from and does nothing.
func Optimize(ctx context.Context, sqlDB *sql.DB) error {
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA optimize"); err != nil {
		return fmt.Errorf("failed to run PRAGMA optimize: %w", err)
	}

	return nil
}

// applyConnectionPragmas runs on every connection the pool opens. The settings
// in "connection.sql" are per-connection in SQLite, so applying them once at
// startup would leave every later connection on SQLite's defaults.
func applyConnectionPragmas(conn *sqlite3.SQLiteConn) error {
	connectionQuery, err := readInitSQL(connectionInitFile)
	if err != nil {
		return fmt.Errorf("failed to read %s script: %w", connectionInitFile, err)
	}

	if _, err := conn.Exec(connectionQuery, nil); err != nil {
		return fmt.Errorf("failed to run connection PRAGMA commands: %w", err)
	}

	return verifyForeignKeys(conn)
}

// verifyForeignKeys fails the connection outright when foreign key enforcement
// did not take, so a silently skipped ON DELETE CASCADE surfaces as an error
// instead of orphaned rows.
func verifyForeignKeys(conn *sqlite3.SQLiteConn) error {
	enabled, err := readPragmaInt(conn, "PRAGMA foreign_keys")
	if err != nil {
		return err
	}

	if enabled != 1 {
		return ErrForeignKeysDisabled
	}

	return nil
}

func readPragmaInt(conn *sqlite3.SQLiteConn, pragma string) (int64, error) {
	rows, err := conn.Query(pragma, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to query %q: %w", pragma, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	values := make([]driver.Value, 1)
	if err := rows.Next(values); err != nil {
		return 0, fmt.Errorf("%q: %w", pragma, ErrPragmaNoValue)
	}

	value, ok := values[0].(int64)
	if !ok {
		return 0, fmt.Errorf("%q: %w", pragma, ErrPragmaNoValue)
	}

	return value, nil
}

func readInitSQL(name string) (string, error) {
	b, err := initPragmas.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
