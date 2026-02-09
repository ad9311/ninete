package db

import (
	"bufio"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/pressly/goose/v3"
)

var ErrEmptyName = errors.New("migration name cannot be empty")

const migrationsPath = "migrations"

//go:embed migrations/*.sql
var embededMigrations embed.FS

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

func CreateMigration() error {
	sqlDB, err := setUpMigrator()
	if err != nil {
		return err
	}

	migrationName, err := promptMigrationName()
	if err != nil {
		return err
	}

	if err := goose.Create(sqlDB, "internal/db/"+migrationsPath, migrationName, "sql"); err != nil {
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
	app, err := prog.Load()
	if err != nil {
		return nil, err
	}

	sqlDB, err := Open()
	if err != nil {
		return sqlDB, err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			app.Logger.Errorf("failed to close database: %v", err)
		}
	}()

	goose.SetBaseFS(embededMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return sqlDB, fmt.Errorf("failed to set dialect: %w", err)
	}

	return sqlDB, nil
}

func promptMigrationName() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Migration name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrEmptyName
	}

	return name, nil
}
