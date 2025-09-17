package logic

import (
	"fmt"
	"os"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-playground/validator/v10"
)

type Store struct {
	app       *prog.App
	queries   repo.Queries
	validate  *validator.Validate
	tokenVars tokenVars
}

type tokenVars struct {
	jwtSecret   string
	jwtIssuer   string
	jwtAudience []string
}

func New(app *prog.App, queries repo.Queries) (*Store, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("'JWT_SECRET' %w", prog.ErrEnvNoTSet)
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		return nil, fmt.Errorf("JWT_ISSUER' %w", prog.ErrEnvNoTSet)
	}

	jwtAudience, err := prog.LoadList("JWT_AUDIENCE")
	if err != nil {
		return nil, err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	return &Store{
		app:      app,
		queries:  queries,
		validate: validate,
		tokenVars: tokenVars{
			jwtSecret:   jwtSecret,
			jwtIssuer:   jwtIssuer,
			jwtAudience: jwtAudience,
		},
	}, nil
}
