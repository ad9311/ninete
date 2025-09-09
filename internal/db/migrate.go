package db

import (
	"database/sql"
	"embed"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/errs"
	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" with database/sql
	"github.com/pressly/goose/v3"
)

const migrationsPath = "migrations"

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunMigrationsUp applies all available database migrations.
func RunMigrationsUp(_ []string) error {
	db, err := setUpMigrator()
	if err != nil {
		return err
	}
	if err := goose.Up(db, migrationsPath); err != nil {
		return err
	}

	return db.Close()
}

// RunMigrationsDown rolls back the most recent migration.
func RunMigrationsDown(_ []string) error {
	db, err := setUpMigrator()
	if err != nil {
		return err
	}
	if err := goose.Down(db, migrationsPath); err != nil {
		return err
	}

	return db.Close()
}

// PrintStatus prints the current status of all database migrations.
func PrintStatus(_ []string) error {
	db, err := setUpMigrator()
	if err != nil {
		return err
	}
	if err := goose.Status(db, migrationsPath); err != nil {
		return err
	}

	return db.Close()
}

// setUpMigrator initializes and returns a database connection for running migrations.
// It loads the application configuration, opens a PostgreSQL database connection using the pgx driver,
// sets up the embedded migration files for Goose, and configures Goose to use the PostgreSQL dialect.
func setUpMigrator() (*sql.DB, error) {
	var db *sql.DB
	config, err := app.LoadConfig()
	if err != nil {
		return db, err
	}

	db, err = sql.Open("pgx", config.DBConfig.URL)
	if err != nil {
		return db, errs.WrapErrorWithMessage("failed to open database", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return db, errs.WrapErrorWithMessage("failed to select database dialect", err)
	}

	return db, nil
}
