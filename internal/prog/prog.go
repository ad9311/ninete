// Package prog provides functionality for loading and validating the application environment,
package prog

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ad9311/ninete/internal/errs"
	"github.com/joho/godotenv"
)

// Environment constants
const (
	ENVProduction  = "production"
	ENVDevelopment = "development"
	ENVTest        = "test"
)

// App represents the main application configuration and dependencies.
type App struct {
	ENV    string
	Logger *Logger
}

// Load initializes the application environment by validating and loading environment variables.
func Load() (*App, error) {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		return nil, fmt.Errorf("%w: application environment", errs.ErrEnvNoTSet)
	}

	if err := isValidENV(env); err != nil {
		return nil, err
	}

	if env != ENVProduction {
		path, ok := findRelativeENVFile()
		if err := godotenv.Load(path); !ok || err != nil {
			return nil, fmt.Errorf("failed to load .env, file %w", err)
		}
	}

	return &App{
		ENV:    env,
		Logger: NewLogger(),
	}, nil
}

// LoadList retrieves the value of the environment variable specified by envName,
func LoadList(envName string) ([]string, error) {
	str := os.Getenv(envName)
	if str == "" {
		return []string{}, fmt.Errorf("%w: %s", errs.ErrEnvNoTSet, envName)
	}

	return strings.Split(str, ","), nil
}

// SetInt retrieves an integer value from the environment variable specified by envName.
// If the environment variable is not set, it returns the provided default value def.
func SetInt(envName string, def int) (int, error) {
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
