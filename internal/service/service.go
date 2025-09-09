// Package service is the layer between database and server
package service

import (
	"regexp"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/repo"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides access to the database pool, SQLC-generated queries,
// application configuration, and request validation.
type Store struct {
	db       *pgxpool.Pool       // Database connection pool
	queries  *repo.Queries       // SQLC-generated query methods
	config   *app.Config         // Application configuration
	validate *validator.Validate // Validator instance for request validation
}

// New initializes and returns a new Store instance with the provided configuration and database pool.
func New(config *app.Config, db *pgxpool.Pool) (*Store, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		re := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

		return re.MatchString(fl.Field().String())
	})
	if err != nil {
		return nil, err
	}

	s := &Store{
		db:       db,
		config:   config,
		queries:  repo.New(db),
		validate: validate,
	}

	return s, nil
}

// Pool returns the database pool for testing and maintenance environments.
func (s *Store) Pool() (*pgxpool.Pool, error) {
	if s.config.IsSafeEnv() {
		return nil, errs.ErrServiceFuncNotAvailable
	}

	return s.db, nil
}

// Queries returns the SQLC-generated Queries for testing and maintenance environments.
func (s *Store) Queries() (*repo.Queries, error) {
	if s.config.IsSafeEnv() {
		return nil, errs.ErrServiceFuncNotAvailable
	}

	return s.queries, nil
}

// ClosePool closes the database connection pool.
func (s *Store) ClosePool() {
	s.db.Close()
}
