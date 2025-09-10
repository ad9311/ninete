package conf

import "os"

// ServerConf
type ServerConf struct {
	Port string
}

// LoadServerConf
func LoadServerConf() (ServerConf, error) {
	var sc ServerConf

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sc = ServerConf{
		Port: port,
	}

	return sc, nil
}
