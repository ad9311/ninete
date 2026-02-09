package prog

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	ENVProduction  = "production"
	ENVDevelopment = "development"
	ENVTest        = "test"
)

type App struct {
	ENV    string
	Logger *Logger
}

func Load() (*App, error) {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		return nil, fmt.Errorf("%w: application environment", ErrEnvNoTSet)
	}

	if err := isValidENV(env); err != nil {
		return nil, err
	}

	if env != ENVProduction {
		path, ok := findRelativeENVFile()
		if err := godotenv.Load(path); !ok || err != nil {
			return nil, fmt.Errorf("failed to load '.env', file %w", err)
		}
	}

	return &App{
		ENV: env,
		Logger: NewLogger(LogOptions{
			EnableColor: env != ENVProduction,
			EnableQuery: true,
		}),
	}, nil
}

func (a *App) IsProduction() bool {
	return a.ENV == ENVProduction
}

func (a *App) IsDevelopment() bool {
	return a.ENV == ENVDevelopment
}

func (a *App) IsTest() bool {
	return a.ENV == ENVTest
}

func LoadList(envName string) ([]string, error) {
	str := os.Getenv(envName)
	if str == "" {
		return []string{}, fmt.Errorf("%w: %s", ErrEnvNoTSet, envName)
	}

	return strings.Split(str, ","), nil
}

func SetInt(envName string, def int) (int, error) {
	envStr := os.Getenv(envName)
	if envStr == "" {
		return def, nil
	}
	v, err := strconv.ParseInt(envStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %w", envStr, err)
	}

	return int(v), nil
}

func isValidENV(env string) error {
	ok := validENVs()[env]
	if !ok {
		return fmt.Errorf("%w '%s'", ErrInvalidEnv, env)
	}

	return nil
}

func validENVs() map[string]bool {
	envs := []string{ENVProduction, ENVDevelopment, ENVTest}
	mappedEnvs := map[string]bool{}

	for _, v := range envs {
		mappedEnvs[v] = true
	}

	return mappedEnvs
}

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

func fileExists(p string) bool {
	_, err := os.Stat(p)

	return err == nil
}
