package app

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ad9311/go-api-base/internal/errs"
)

const (
	defaultMaxConns = 20
	defaultMinConns = 5

	defaultMaxConnIdleTime = 5 * time.Minute
	defaultMaxConnLifetime = 30 * time.Minute
)

// DBConfig holds the configuration settings required to connect to the database,
type DBConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnIdleTime time.Duration
	MaxConnLifetime time.Duration
}

// setDBConfig initializes and returns a DBConfig struct based on the provided environment.
func setDBConfig(env string) (DBConfig, error) {
	var dbConfig DBConfig

	url, err := buildDBURL(env)
	if err != nil {
		return dbConfig, err
	}

	maxConns, err := setCount("MAX_CONNS", defaultMaxConns)
	if err != nil {
		return dbConfig, err
	}

	minConns, err := setCount("MIN_CONNS", defaultMinConns)
	if err != nil {
		return dbConfig, err
	}

	maxConnIdleTime, err := setDuration("MAX_CONN_IDLE_TIME", defaultMaxConnIdleTime)
	if err != nil {
		return dbConfig, err
	}

	maxConnLifetime, err := setDuration("MAX_CONN_LIFETIME", defaultMaxConnLifetime)
	if err != nil {
		return dbConfig, err
	}

	dbConfig = DBConfig{
		URL:             url,
		MaxConns:        maxConns,
		MinConns:        minConns,
		MaxConnIdleTime: maxConnIdleTime,
		MaxConnLifetime: maxConnLifetime,
	}

	return dbConfig, nil
}

// buildDBURL constructs the database connection URL based on the environment and environment variables.
func buildDBURL(env string) (string, error) {
	if envURL, ok := os.LookupEnv("DATABASE_URL"); ok && envURL != "" {
		return envURL, nil
	}

	var prefix string
	if env == EnvTest {
		prefix = "TEST_"
	}

	user := os.Getenv(prefix + "DB_USER")
	password := os.Getenv(prefix + "DB_PASSWORD")
	port := os.Getenv(prefix + "DB_PORT")
	name := os.Getenv(prefix + "DB_NAME")

	params := parseRuntimeParams(prefix)

	if slices.Contains([]string{user, password, port, name}, "") {
		return "", errs.ErrDatabaseVarsNotSet
	}

	url := fmt.Sprintf(
		"postgresql://%s:%s@localhost:%s/%s%s",
		user,
		password,
		port,
		name,
		params,
	)

	return url, nil
}

// parseRuntimeParams constructs the runtime parameters string for the database URL.
// It reads the environment variable for additional parameters, replaces ":" with "="
// and spaces with "&" to format them as URL query parameters.
func parseRuntimeParams(prefix string) string {
	sslMode := "?sslmode=disable"

	params := os.Getenv(prefix + "RUNTIME_PARAMS")

	if params == "" {
		return sslMode
	}

	params = strings.ReplaceAll(params, ":", "=")
	params = strings.ReplaceAll(params, " ", "&")

	return sslMode + "&" + params
}

// setCount retrieves an integer value from the environment variable specified by envName.
// If the variable is not set, it returns the provided default value (def).
// Returns an error if the environment variable is set but cannot be parsed as an int32.
func setCount(envName string, def int32) (int32, error) {
	envValue := os.Getenv(envName)
	if envValue == "" {
		return def, nil
	}

	value, err := strconv.ParseInt(envValue, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(value), nil
}

// setDuration retrieves the value of the environment variable specified by envName,
// parses it as a time.Duration, and returns the result. If the environment variable
// is not set, it returns the provided default duration def. If the value cannot be
// parsed as a valid duration, an error is returned.
//
// Example environment variable value: "30s", "1m", "2h45m".
// See time.ParseDuration for supported formats.
func setDuration(envName string, def time.Duration) (time.Duration, error) {
	envValue := os.Getenv(envName)
	if envValue == "" {
		return def, nil
	}

	value, err := time.ParseDuration(envValue)
	if err != nil {
		return 0, err
	}

	return value, nil
}
