package conf

import (
	"os"
	"strconv"
)

const (
	defaultMaxOpenConns = 1
	defaultMaxIdleConns = 1
)

// DBConf
type DBConf struct {
	URL          string
	MaxOpenConns int
	MaxIdleConns int
}

// LoadDBConf
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
