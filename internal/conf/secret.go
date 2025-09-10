package conf

import "os"

// Secrets holds sensitive configuration values required by the application,
// such as cryptographic keys and tokens.
type Secrets struct {
	JWTSecret string
}

// LoadSecrets loads secret configuration values from environment variables.
func LoadSecrets() (Secrets, error) {
	var scrt Secrets

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return scrt, nil // ERROR
	}

	scrt = Secrets{
		JWTSecret: jwtSecret,
	}

	return scrt, nil
}
