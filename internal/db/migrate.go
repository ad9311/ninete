package db

import (
	"database/sql"
	"embed"
	"log"

	"github.com/ad9311/ninete/internal/conf"
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
	defer closeDB(sqlDB)

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
	defer closeDB(sqlDB)

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
	defer closeDB(sqlDB)

	if err := goose.Status(sqlDB, migrationsPath); err != nil {
		return err
	}

	return nil
}

// setUpMigrator initializes and returns a database connection for running migrations.
// It loads the application configuration, opens a PostgreSQL database connection using the pgx driver,
// sets up the embedded migration files for Goose, and configures Goose to use the PostgreSQL dialect.
func setUpMigrator() (*sql.DB, error) {
	var sqlDB *sql.DB

	ac, err := conf.Load()
	if err != nil {
		return sqlDB, err
	}

	sqlDB, err = Open(ac.DBConf)
	if err != nil {
		return sqlDB, err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return sqlDB, err
	}

	return sqlDB, nil
}

// closeDB attempts to close the provided sql.DB connection.
// If an error occurs during the close operation, it logs the error.
// This function helps ensure that database resources are properly released.
func closeDB(sqlDB *sql.DB) {
	if err := sqlDB.Close(); err != nil {
		log.Println(err)
	}
}
