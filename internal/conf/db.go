package conf

import "os"

// DBConf
type DBConf struct {
	URL string
}

// LoadDBConf
func LoadDBConf() (DBConf, error) {
	var dbc DBConf

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return dbc, nil // ERROR
	}

	dbc = DBConf{
		URL: url,
	}

	return dbc, nil
}
