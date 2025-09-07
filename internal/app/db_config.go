package app

import (
	"fmt"
	"os"
	"slices"

	"github.com/ad9311/go-api-base/internal/errs"
)

// DBConfig holds the configuration settings required to connect to the database,
type DBConfig struct {
	URL            string
	MigrationsPath string
}

// setDBConfig initializes and returns a DBConfig struct based on the provided environment.
func setDBConfig(env string) (DBConfig, error) {
	var dbConfig DBConfig

	url, err := buildDBURL(env)
	if err != nil {
		return dbConfig, err
	}

	migPath := os.Getenv("MIGRATIONS_PATH")
	if migPath == "" {
		return dbConfig, errs.ErrMigrationPath
	}

	dbConfig = DBConfig{
		URL:            url,
		MigrationsPath: migPath,
	}

	return dbConfig, nil
}

// buildDBURL constructs the database connection URL based on the environment and environment variables.
func buildDBURL(env string) (string, error) {
	var prefix string
	if env == EnvTest {
		prefix = "TEST_"
	}

	user := os.Getenv(prefix + "DB_USER")
	password := os.Getenv(prefix + "DB_PASSWORD")
	port := os.Getenv(prefix + "DB_PORT")
	name := os.Getenv(prefix + "DB_NAME")

	if slices.Contains([]string{user, password, port, name}, "") {
		return "", errs.ErrDatabaseVarsNotSet
	}

	url := fmt.Sprintf(
		"postgresql://%s:%s@localhost:%s/%s?sslmode=disable",
		user,
		password,
		port,
		name,
	)

	return url, nil
}
