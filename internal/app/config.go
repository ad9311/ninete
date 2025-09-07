package app

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

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

// DefaultTimeout is the default value for context timeouts
const DefaultTimeout = 2 * time.Second

const (
	envFile                = ".env"
	envVar                 = "ENV"
	defaultMaxConns        = int32(20)
	defaultMinConns        = int32(5)
	defaultMaxConnLifeTime = 30 * time.Second
	defaultMaxConnIdleTime = 5 * time.Second
	defaultPort            = "8080"
)

// Config holds the application's runtime configuration populated from
// environment variables (and a .env file when applicable).
type Config struct {
	Logger          *console.Logger // logger is the application's logger instance (internal use only)
	Env             string          // Env is the environment in which the app is running (production, development, test, maintenance)
	Port            string          // Port is the port the server listens on
	DBURL           string          // DBURL is the database connection URL
	MigrationsPath  string          // MigrationsPath is the path to the database migrations directory
	MaxConns        int32           // MaxConns is the maximum number of open connections for the database pool
	MinConns        int32           // MinConns is the minimum number of open connections for the database pool
	MaxConnIdleTime time.Duration   // MaxConnIdleTime is the maximum duration a connection may be idle before being closed
	MaxConnLifetime time.Duration   // MaxConnLifetime is the maximum total duration a connection may be reused before being closed
	JWTSecret       []byte          // JWTSecret is the secret used to sign JWT access tokens
	JWTIssuer       string          // JWTIssuer is the issuer claim to set in JWT tokens
	JWTAudience     []string        // JWTAudience is the audience claim to set in JWT tokens
	AllowedOrigins  []string        // AllowedOrigins is the list of allowed CORS origins for the server
}

// LoadConfig loads the app configuration from environment variables. It will
// load a .env file (from a parent directory) when the environment is not
// production and SKIP_ENV_FILE is not set.
func LoadConfig() (*Config, error) {
	env, err := loadEnv()
	if err != nil {
		return nil, err
	}

	dbURL, err := buildDBURL(env)
	if err != nil {
		return nil, err
	}

	migPath := os.Getenv("MIGRATIONS_PATH")
	if migPath == "" {
		return nil, errs.ErrMigrationPath
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
		Logger:          logger,
		Env:             env,
		DBURL:           dbURL,
		Port:            port,
		MigrationsPath:  migPath,
		MaxConns:        defaultMaxConns,
		MinConns:        defaultMinConns,
		MaxConnIdleTime: defaultMaxConnIdleTime,
		MaxConnLifetime: defaultMaxConnLifeTime,
		JWTSecret:       []byte(jwtSecret),
		JWTIssuer:       jwtIssuer,
		JWTAudience:     jwtAudience,
		AllowedOrigins:  allowedOrigns,
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

// buildDBURL constructs the database connection URL based on the environment and environment variables.
func buildDBURL(env string) (string, error) {
	var prefix string
	if env == EnvTest {
		prefix = "TEST_"
	}

	user := os.Getenv(prefix + "DB_USER")
	password := os.Getenv(prefix + "DB_PASSWORD")
	port := os.Getenv(prefix + "DB_PORT")
	name := os.Getenv(prefix + "DB_NAME")

	if slices.Contains([]string{user, password, port, name}, "") {
		return "", errs.ErrDatabaseVarsNotSet
	}

	url := fmt.Sprintf(
		"postgresql://%s:%s@localhost:%s/%s?sslmode=disable&connect_timeout=5&application_name=go_api_base",
		user,
		password,
		port,
		name,
	)

	return url, nil
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
