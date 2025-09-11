package db

import (
	"embed"

	"github.com/ad9311/ninete/internal/conf"
	"github.com/pressly/goose/v3"
)

const migrationsPath = "migrations"

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunMigrationsUp applies all available database migrations.
func RunMigrationsUp() error {
	conn, err := setUpMigrator()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := goose.Up(conn.DB, migrationsPath); err != nil {
		return err
	}

	return nil
}

// RunMigrationsDown rolls back the most recent migration.
func RunMigrationsDown() error {
	conn, err := setUpMigrator()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := goose.Down(conn.DB, migrationsPath); err != nil {
		return err
	}

	return nil
}

// PrintStatus prints the current status of all database migrations.
func PrintStatus() error {
	conn, err := setUpMigrator()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := goose.Status(conn.DB, migrationsPath); err != nil {
		return err
	}

	return nil
}

// setUpMigrator initializes and returns a database connection for running migrations.
// It loads the application configuration, opens a PostgreSQL database connection using the pgx driver,
// sets up the embedded migration files for Goose, and configures Goose to use the PostgreSQL dialect.
func setUpMigrator() (*Pool, error) {
	var conn *Pool

	ac, err := conf.Load()
	if err != nil {
		return conn, err
	}

	conn, err = Open(ac.DBConf)
	if err != nil {
		return conn, err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return conn, err
	}

	return conn, nil
}
