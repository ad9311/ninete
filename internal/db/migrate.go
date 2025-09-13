package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/pressly/goose/v3"
)

const migrationsPath = "migrations"

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrationsUp() error {
	sqlDB, err := setUpMigrator()
	if err != nil {
		return err
	}

	if err := goose.Up(sqlDB, migrationsPath); err != nil {
		return err
	}

	if err := sqlDB.Close(); err != nil {
		return err
	}

	return nil
}

func RunMigrationsDown() error {
	sqlDB, err := setUpMigrator()
	if err != nil {
		return err
	}

	if err := goose.Down(sqlDB, migrationsPath); err != nil {
		return err
	}

	if err := sqlDB.Close(); err != nil {
		return err
	}

	return nil
}

func PrintStatus() error {
	sqlDB, err := setUpMigrator()
	if err != nil {
		return err
	}

	if err := goose.Status(sqlDB, migrationsPath); err != nil {
		return err
	}

	if err := sqlDB.Close(); err != nil {
		return err
	}

	return nil
}

func setUpMigrator() (*sql.DB, error) {
	_, err := prog.Load()
	if err != nil {
		return nil, err
	}

	sqlDB, err := Open()
	if err != nil {
		return sqlDB, err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return sqlDB, fmt.Errorf("failed to set dialect: %w", err)
	}

	return sqlDB, nil
}
