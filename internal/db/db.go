package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"

	"github.com/ad9311/ninete/internal/prog"
	_ "github.com/mattn/go-sqlite3" // Database driver
)

const (
	DefaultMaxOpenConns = 1
	DefaultMaxIdleConns = 1
)

const initFile = "init/init.sql"

//go:embed init/*.sql
var initPragmas embed.FS

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

	sqlDB, err := sql.Open("sqlite3", "file:"+url+"?_loc=UTC")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	initQuery, err := readInitSQL(initFile)
	if err != nil {
		return nil, fmt.Errorf("failed to run %s script: %w", initFile, err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err := sqlDB.Exec(initQuery); err != nil {
		return nil, fmt.Errorf("failed to run init PRAGMA commands: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)

	return sqlDB, nil
}

func readInitSQL(name string) (string, error) {
	b, err := initPragmas.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
