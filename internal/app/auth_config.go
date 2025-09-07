package app

import (
	"os"
	"strings"

	"github.com/ad9311/go-api-base/internal/errs"
)

// AuthConfig holds configuration settings related to authentication, including
// JWT token signing, issuer and audience claims, and allowed CORS origins.
type AuthConfig struct {
	JWTSecret      []byte   // JWTSecret is the secret used to sign JWT access tokens
	JWTIssuer      string   // JWTIssuer is the issuer claim to set in JWT tokens
	JWTAudience    []string // JWTAudience is the audience claim to set in JWT tokens
	AllowedOrigins []string // AllowedOrigins is the list of allowed CORS origins for the server
}

// setAuthConfig reads authentication-related configuration from environment variables,
// validates them, and returns an AuthConfig struct. It returns an error if any required
// configuration is missing or invalid.
func setAuthConfig() (AuthConfig, error) {
	var authConfig AuthConfig

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return authConfig, errs.ErrJWTSecretNotSet
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		return authConfig, errs.ErrJWTIssuerNotSet
	}

	jwtAudienceValue := os.Getenv("JWT_AUDIENCE")
	if jwtAudienceValue == "" {
		return authConfig, errs.ErrJWTAudienceNotSet
	}
	jwtAudience := parseValueList(jwtAudienceValue)
	if len(jwtAudience) == 0 {
		return authConfig, errs.ErrJWTAudienceNotSet
	}

	allowedOrignsValue := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrignsValue == "" {
		return authConfig, errs.ErrAllowedOriginsNotSet
	}
	allowedOrigns := parseValueList(allowedOrignsValue)
	if len(allowedOrigns) == 0 {
		return authConfig, errs.ErrAllowedOriginsNotSet
	}

	authConfig = AuthConfig{
		JWTSecret:      []byte(jwtSecret),
		JWTIssuer:      jwtIssuer,
		JWTAudience:    jwtAudience,
		AllowedOrigins: allowedOrigns,
	}

	return authConfig, nil
}

// parseValueList splits the input string by commas and returns a slice of substrings.
// It is useful for parsing comma-separated lists from configuration values.
func parseValueList(list string) []string {
	return strings.Split(list, ",")
}
