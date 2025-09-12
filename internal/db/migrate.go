package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/ad9311/ninete/internal/app"
	"github.com/pressly/goose/v3"
)

const migrationsPath = "migrations"

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunMigrationsUp applies all available database migrations.
func RunMigrationsUp() error {
	sqlDB, err := setUpMigrator()
	if err != nil {
		return err
	}

	if err := goose.Up(sqlDB, migrationsPath); err != nil {
		return err
	}

	return nil
}

// RunMigrationsDown rolls back the most recent migration.
func RunMigrationsDown() error {
	sqlDB, err := setUpMigrator()
	if err != nil {
		return err
	}

	if err := goose.Down(sqlDB, migrationsPath); err != nil {
		return err
	}

	return nil
}

// PrintStatus prints the current status of all database migrations.
func PrintStatus() error {
	sqlDB, err := setUpMigrator()
	if err != nil {
		return err
	}

	if err := goose.Status(sqlDB, migrationsPath); err != nil {
		return err
	}

	return nil
}

// setUpMigrator initializes and returns a database connection for running migrations.
func setUpMigrator() (*sql.DB, error) {
	var sqlDB *sql.DB

	_, err := app.Load()
	if err != nil {
		return sqlDB, err
	}

	sqlDB, err = Open()
	if err != nil {
		return sqlDB, err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return sqlDB, fmt.Errorf("failed to set dialect: %w", err)
	}

	return sqlDB, nil
}
