package app

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ad9311/go-api-base/internal/console"
	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/joho/godotenv"
)

// Environment variables.
// If a new one is added it should also be added in the getValidEnvs function.
const (
	EnvProduction  = "production"
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvMaintenance = "maintenance"
)

const (
	envFile     = ".env"
	envVar      = "ENV"
	defaultPort = "8080"
)

// Config holds the application's runtime configuration populated from
// environment variables (and a .env file when applicable).
type Config struct {
	Env            string   // Env is the environment in which the app is running (production, development, test, maintenance)
	Port           string   // Port is the port the server listens on
	JWTSecret      []byte   // JWTSecret is the secret used to sign JWT access tokens
	JWTIssuer      string   // JWTIssuer is the issuer claim to set in JWT tokens
	JWTAudience    []string // JWTAudience is the audience claim to set in JWT tokens
	AllowedOrigins []string // AllowedOrigins is the list of allowed CORS origins for the server
	DBConfig       DBConfig
	Logger         *console.Logger // logger is the application's logger instance
}

// LoadConfig loads the app configuration from environment variables. It will
// load a .env file (from a parent directory) when the environment is not
// production and SKIP_ENV_FILE is not set.
func LoadConfig() (*Config, error) {
	env, err := loadEnv()
	if err != nil {
		return nil, err
	}

	dbConfig, err := setDBConfig(env)
	if err != nil {
		return nil, err
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errs.ErrJWTSecretNotSet
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		return nil, errs.ErrJWTIssuerNotSet
	}

	jwtAudienceValue := os.Getenv("JWT_AUDIENCE")
	if jwtAudienceValue == "" {
		return nil, errs.ErrJWTAudienceNotSet
	}
	jwtAudience := parseValueList(jwtAudienceValue)
	if len(jwtAudience) == 0 {
		return nil, errs.ErrJWTAudienceNotSet
	}

	allowedOrignsValue := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrignsValue == "" {
		return nil, errs.ErrAllowedOriginsNotSet
	}
	allowedOrigns := parseValueList(allowedOrignsValue)
	if len(allowedOrigns) == 0 {
		return nil, errs.ErrAllowedOriginsNotSet
	}

	logger := console.New(nil, nil, env != EnvProduction)

	return &Config{
		Env:            env,
		Port:           port,
		JWTSecret:      []byte(jwtSecret),
		JWTIssuer:      jwtIssuer,
		JWTAudience:    jwtAudience,
		AllowedOrigins: allowedOrigns,
		DBConfig:       dbConfig,
		Logger:         logger,
	}, nil
}

// IsSafeEnv returns true when the current environment is considered a
// "safe" runtime for normal operation. The function currently treats
// development and production as safe environments.
func (c *Config) IsSafeEnv() bool {
	if c.Env == EnvDevelopment || c.Env == EnvProduction {
		return true
	}

	return false
}

// loadEnv loads the environment variable and .env file if needed, returning the environment name or an error.
func loadEnv() (string, error) {
	env, ok := os.LookupEnv(envVar)
	if !ok {
		return "", errs.ErrNoEnv
	}
	validEnvs := getValidEnvs()
	if !validEnvs[env] {
		var envs []string
		for k := range validEnvs {
			envs = append(envs, k)
		}
		slices.Sort(envs)

		return "", errs.WrapMessageWithError(errs.ErrInvalidEnv, "must be one of "+strings.Join(envs, ", "))
	}

	if env != EnvProduction && os.Getenv("SKIP_ENV_FILE") == "" {
		path, ok := findRelativeENVFile()
		err := godotenv.Load(path)

		if !ok || err != nil {
			return "", errs.ErrEnvLoad
		}
	}

	return env, nil
}

// getValidEnvs returns a list of valid environment names.
func getValidEnvs() map[string]bool {
	return map[string]bool{
		EnvDevelopment: true,
		EnvProduction:  true,
		EnvTest:        true,
		EnvMaintenance: true,
	}
}

// findRelativeENVFile searches for a .env file in the current or parent directories and returns its path if found.
func findRelativeENVFile() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		p := filepath.Join(dir, envFile)
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

// fileExists checks if the file at the given path exists.
func fileExists(p string) bool {
	_, err := os.Stat(p)

	return err == nil
}

// parseValueList splits the input string by commas and returns a slice of substrings.
// It is useful for parsing comma-separated lists from configuration values.
func parseValueList(list string) []string {
	return strings.Split(list, ",")
}
