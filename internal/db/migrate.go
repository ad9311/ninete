package db

import (
	"errors"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // For connection to postgres
	_ "github.com/golang-migrate/migrate/v4/source/file"       // For reading migration files
)

// RunMigrationsUp loads configuration, creates a migration instance, and runs all pending migrations up.
func RunMigrationsUp(_ []string) error {
	config, err := app.LoadConfig()
	if err != nil {
		return err
	}

	migrator, err := createMigrator(config)
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		if err := closeMigrator(migrator); err != nil {
			return err
		}

		return errs.WrapErrorWithMessage("failed to run migrations down", err)
	}

	if err := closeMigrator(migrator); err != nil {
		return err
	}

	return nil
}

// RunMigrationsDown loads configuration, creates a migration instance, and runs one migration down.
func RunMigrationsDown(_ []string) error {
	config, err := app.LoadConfig()
	if err != nil {
		return err
	}

	migrator, err := createMigrator(config)
	if err != nil {
		return err
	}

	if err := migrator.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		if err := closeMigrator(migrator); err != nil {
			return err
		}

		return errs.WrapErrorWithMessage("failed to run migrations down", err)
	}

	if err := closeMigrator(migrator); err != nil {
		return err
	}

	return nil
}

// createMigrator creates a new migrate.Migrate instance using the database URL and migrations path.
func createMigrator(config *app.Config) (*migrate.Migrate, error) {
	var migrator *migrate.Migrate

	if config.DBConfig.URL == "" {
		return migrator, errs.ErrDatabaseVarsNotSet
	}

	migrator, err := migrate.New("file://"+config.DBConfig.MigrationsPath, config.DBConfig.URL)
	if err != nil {
		return migrator, errs.WrapErrorWithMessage("failed to create migrate instance", err)
	}

	return migrator, nil
}

// closeMigrator closes the migration instance and joins any errors from source and database.
func closeMigrator(m *migrate.Migrate) error {
	sourceErr, dbErr := m.Close()

	return errors.Join(sourceErr, dbErr)
}
