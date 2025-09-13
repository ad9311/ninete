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
	initFile = "init/init.sql"

	defaultMaxOpenConns = 1
	defaultMaxIdleConns = 1
)

//go:embed init/*.sql
var initPragmas embed.FS

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

	initQuery, err := readInitSQL(initFile)
	if err != nil {
		return nil, fmt.Errorf("failed to run %s script: %w", initFile, err)
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
