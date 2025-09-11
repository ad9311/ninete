package conf

import (
	"fmt"
	"os"

	"github.com/ad9311/ninete/internal/errs"
)

// Secrets holds sensitive configuration values required by the application.
type Secrets struct {
	JWTSecret string
}

// LoadSecrets loads secret configuration values from environment variables.
func LoadSecrets() (Secrets, error) {
	var scrt Secrets

	envName := "JWT_SECRET"
	jwtSecret := os.Getenv(envName)
	if jwtSecret == "" {
		return scrt, fmt.Errorf("%w: %s", errs.ErrEnvNoTSet, envName)
	}

	scrt = Secrets{
		JWTSecret: jwtSecret,
	}

	return scrt, nil
}
