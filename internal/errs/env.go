package errs

import "errors"

// Environment/config errors
var (
	ErrDatabaseVarsNotSet   = errors.New("variable DATABASE variables not set")
	ErrJWTSecretNotSet      = errors.New("variable JWT_SECRET not set")
	ErrJWTIssuerNotSet      = errors.New("variable JWT_ISSUER not set")
	ErrJWTAudienceNotSet    = errors.New("variable JWT_AUDIENCE not set")
	ErrAllowedOriginsNotSet = errors.New("variable ALLOWED_ORIGINS not set")
	ErrMigrationPath        = errors.New("variable MIGRATIONS_PATH not set")
	ErrNoEnv                = errors.New("variable ENV not set")
	ErrInvalidEnv           = errors.New("invalid ENV")
	ErrEnvLoad              = errors.New("failed to load .env file")
)
