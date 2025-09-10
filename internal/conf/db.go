package conf

import (
	"os"
	"strconv"
)

// Database connection pool configuration constants.
const (
	defaultMaxOpenConns = 1
	defaultMaxIdleConns = 1
)

// DBConf holds the configuration parameters for the database connection.
type DBConf struct {
	URL          string
	MaxOpenConns int
	MaxIdleConns int
}

// LoadDBConf loads the database configuration from environment variables.
// Returns a DBConf struct populated with these values, or an error if any
// configuration value is invalid or cannot be parsed.
func LoadDBConf() (DBConf, error) {
	var dbc DBConf

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return dbc, nil // ERROR
	}

	maxOpenConns, err := setInt("MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return dbc, err // ERROR
	}

	maxIdleConns, err := setInt("MAX_IDLE_CONNS", defaultMaxIdleConns)
	if err != nil {
		return dbc, err // ERROR
	}

	dbc = DBConf{
		URL:          url,
		MaxOpenConns: maxOpenConns,
		MaxIdleConns: maxIdleConns,
	}

	return dbc, nil
}

// setInt retrieves an integer value from the environment variable specified by envName.
// If the environment variable is not set, it returns the provided default value def.
func setInt(envName string, def int) (int, error) {
	maxConnsStr := os.Getenv(envName)
	if maxConnsStr == "" {
		return def, nil
	}
	v, err := strconv.ParseInt(maxConnsStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}
