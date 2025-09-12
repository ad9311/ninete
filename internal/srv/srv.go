// Package srv provides core service logic and dependencies for the application.
package srv

import (
	"fmt"
	"os"

	"github.com/ad9311/ninete/internal/errs"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

// Store holds the core dependencies and configuration for the service layer.
type Store struct {
	app         *prog.App
	queries     repo.Queries
	jwtSecret   string
	jwtIssuer   string
	jwtAudience []string
}

// New initializes a new Store with required dependencies and configuration loaded from environment variables.
// It returns an error if any required environment variable is missing or invalid.
func New(app *prog.App, queries repo.Queries) (*Store, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("%w: JWT_SECRET", errs.ErrEnvNoTSet)
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		return nil, fmt.Errorf("%w: JWT_ISSUER", errs.ErrEnvNoTSet)
	}

	jwtAudience, err := prog.LoadList("JWT_AUDIENCE")
	if err != nil {
		return nil, err
	}

	return &Store{
		app:         app,
		queries:     queries,
		jwtSecret:   jwtSecret,
		jwtIssuer:   jwtIssuer,
		jwtAudience: jwtAudience,
	}, nil
}
