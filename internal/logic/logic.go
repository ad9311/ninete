package logic

import (
	"fmt"
	"os"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

type Store struct {
	app         *prog.App
	queries     repo.Queries
	jwtSecret   string
	jwtIssuer   string
	jwtAudience []string
}

func New(app *prog.App, queries repo.Queries) (*Store, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("%w: JWT_SECRET", prog.ErrEnvNoTSet)
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		return nil, fmt.Errorf("%w: JWT_ISSUER", prog.ErrEnvNoTSet)
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
