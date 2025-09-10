// Package conf provides functionality for loading and managing application configuration.
package conf

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Environment constants
const (
	ENVProduction  = "production"
	ENVDevelopment = "development"
	ENVTest        = "test"
)

// AppConf holds the main application configuration, including environment settings,
// database configuration, server configuration, and application secrets.
type AppConf struct {
	ENV        string
	DBConf     DBConf
	ServerConf ServerConf
	Secrets    Secrets
}

// Load initializes and returns the application configuration by loading environment variables,
// database configuration, server configuration, and secrets. It returns an AppConf struct
// populated with these values, or an error if any of the configuration loading steps fail.
func Load() (AppConf, error) {
	var ac AppConf

	env, err := loadENV()
	if err != nil {
		return ac, err
	}

	dbc, err := LoadDBConf()
	if err != nil {
		return ac, err
	}

	sc, err := LoadServerConf()
	if err != nil {
		return ac, err
	}

	scrt, err := LoadSecrets()
	if err != nil {
		return ac, err
	}

	ac = AppConf{
		ENV:        env,
		DBConf:     dbc,
		ServerConf: sc,
		Secrets:    scrt,
	}

	return ac, nil
}

// loadENV loads the application environment from the "NINETE_ENV" environment variable.
// If the variable is not set, it returns the empty string and no error.
func loadENV() (string, error) {
	env, ok := os.LookupEnv("NINETE_ENV")
	if !ok {
		return env, nil // ERROR
	}

	if err := isValidENV(env); err != nil {
		return "", err
	}

	if env != ENVProduction {
		path, ok := findRelativeENVFile()
		if err := godotenv.Load(path); !ok || err != nil {
			return "", err
		}
	}

	return env, nil
}

// isValidENV checks if the provided environment string is valid according to the set of allowed environment values.
// It returns an error if the environment is not valid, otherwise returns nil.
func isValidENV(env string) error {
	ok := validENVs()[env]
	if !ok {
		return nil // Error
	}

	return nil
}

// validENVs returns a map indicating the valid environment names for the application.
// The keys are environment constants (e.g., ENVProduction, ENVDevelopment, ENVTest),
// and the values are set to true to signify their validity.
func validENVs() map[string]bool {
	return map[string]bool{
		ENVProduction:  true,
		ENVDevelopment: true,
		ENVTest:        true,
	}
}

// findRelativeENVFile searches for a ".env" file starting from the current working directory
// and traversing up the directory tree. It returns the path to the first ".env" file found and
// a boolean indicating whether the file was found. If no ".env" file is found, it returns an
// empty string and false.
func findRelativeENVFile() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		p := filepath.Join(dir, ".env")
		if fileExists(p) {
			return p, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// fileExists checks whether the file at the given path exists.
// It returns true if the file exists, and false otherwise.
func fileExists(p string) bool {
	_, err := os.Stat(p)

	return err == nil
}
