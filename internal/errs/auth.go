package errs

import "errors"

// JWT/token/auth errors
var (
	ErrInvalidJWTToken      = errors.New("invalid jwt token")
	ErrInvalidJWTIssuer     = errors.New("invalid issuer for jwt token")
	ErrInvalidJWTAudience   = errors.New("invalid audience for jwt token")
	ErrInvalidClaimsType    = errors.New("invalid claims type")
	ErrInvalidAuthHeader    = errors.New("invalid Authorization header")
	ErrRefreshTokenNotFound = errors.New("refresh token cookie not found")
	ErrExpiredRefreshToken  = errors.New("expired refresh token")
	ErrRevokedRefreshToken  = errors.New("revoked refresh token")
	ErrGenerateJWTToken     = errors.New("could not generate JWT token")
)

// UUID/ID errors
var (
	ErrInvalidUUIDFormat = errors.New("invalid UUID format")
	ErrInvalidUUIDHex    = errors.New("invalid UUID hex")
	ErrInvalidUUIDLength = errors.New("invalid UUID length")
	ErrUUIDIsNull        = errors.New("uuid is null")
)
