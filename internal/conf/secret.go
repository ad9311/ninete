package conf

import "os"

// Secrets
type Secrets struct {
	JWTSecret string
}

// LoadSecrets
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
