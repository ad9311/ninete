// Package app provides functionality for loading and validating the application environment,
package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ad9311/ninete/internal/errs"
	"github.com/joho/godotenv"
)

// Environment constants
const (
	ENVProduction  = "production"
	ENVDevelopment = "development"
	ENVTest        = "test"
)

// Load loads the application environment by calling loadENV.
// It returns the loaded environment as a string and any error encountered during loading.
func Load() (string, error) {
	env, err := loadENV()
	if err != nil {
		return env, err
	}

	return env, nil
}

// loadENV loads the application environment from the "ENV" environment variable.
// If the variable is not set, it returns the empty string and no error.
func loadENV() (string, error) {
	envName := "ENV"
	env, ok := os.LookupEnv(envName)
	if !ok {
		return "", fmt.Errorf("%w: %s", errs.ErrEnvNoTSet, envName)
	}

	if err := isValidENV(env); err != nil {
		return "", err
	}

	if env != ENVProduction {
		path, ok := findRelativeENVFile()
		if err := godotenv.Load(path); !ok || err != nil {
			return "", fmt.Errorf("failed to load .env, file %w", err)
		}
	}

	return env, nil
}

// isValidENV checks if the provided environment string is valid.
func isValidENV(env string) error {
	ok := validENVs()[env]
	if !ok {
		return fmt.Errorf("%w: %s", errs.ErrInvalidEnv, env)
	}

	return nil
}

// validENVs returns a map indicating the valid environment names for the application.
func validENVs() map[string]bool {
	return map[string]bool{
		ENVProduction:  true,
		ENVDevelopment: true,
		ENVTest:        true,
	}
}

// findRelativeENVFile searches for a ".env" file starting from the current working directory
// and traversing up the directory tree. If found it returns true, otherwise false.
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
func fileExists(p string) bool {
	_, err := os.Stat(p)

	return err == nil
}
