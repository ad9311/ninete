package conf

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ad9311/ninete/internal/errs"
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
func LoadDBConf() (DBConf, error) {
	var dbc DBConf

	envName := "DATABASE_URL"
	url := os.Getenv(envName)
	if url == "" {
		return dbc, fmt.Errorf("%w: %s", errs.ErrEnvNoTSet, envName)
	}

	maxOpenConns, err := setInt("MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return dbc, err
	}

	maxIdleConns, err := setInt("MAX_IDLE_CONNS", defaultMaxIdleConns)
	if err != nil {
		return dbc, err
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
		return 0, fmt.Errorf("failed to parse %s: %w", maxConnsStr, err)
	}

	return int(v), nil
}
