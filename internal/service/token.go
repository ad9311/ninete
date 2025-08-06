package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/repo"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Token represents either an access token or a refresh token, including its value and timestamps.
type Token struct {
	Value     string    `json:"value"`     // Token string value
	IssuedAt  time.Time `json:"issuedAt"`  // Time the token was issued
	ExpiresAt time.Time `json:"expiresAt"` // Time the token expires
}

// Expiration durations for access and refresh tokens.
const (
	AccessTokenExp  = 15 * time.Minute
	RefreshTokenExp = 7 * 24 * time.Hour
)

// SaveRefreshToken generates and saves a new refresh token for the specified user in the database.
func (s *Store) SaveRefreshToken(ctx context.Context, userID int32) (repo.RefreshToken, error) {
	iat, exp := generateDateClaims(RefreshTokenExp)

	refreshToken, err := s.queries.InsertRefreshToken(ctx, repo.InsertRefreshTokenParams{
		UserID: userID,
		IssuedAt: pgtype.Timestamptz{
			Time:  iat,
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  exp,
			Valid: true,
		},
	})
	if err != nil {
		return refreshToken, errs.HandlePgError(err)
	}

	return refreshToken, nil
}

// FindRefreshTokenByUUID retrieves a refresh token from the database by its UUID string.
func (s *Store) FindRefreshTokenByUUID(ctx context.Context, uuidString string) (repo.RefreshToken, error) {
	var refreshToken repo.RefreshToken

	uuid, err := ParseUUID(uuidString)
	if err != nil {
		return refreshToken, err
	}

	pgUUID := pgtype.UUID{
		Bytes: uuid,
		Valid: true,
	}

	refreshToken, err = s.queries.SelectRefreshTokenByUUID(ctx, pgUUID)
	if err != nil {
		return refreshToken, errs.HandlePgError(err)
	}

	return refreshToken, nil
}

// DeleteExpieredRefreshTokens deletes all expired refresh tokens from the database.
func (s *Store) DeleteExpieredRefreshTokens(ctx context.Context) (int64, error) {
	if s.config.IsSafeEnv() {
		return 0, errs.ErrServiceFuncNotAvailable
	}

	count, err := s.queries.DeleteRefreshTokensWhereExpired(ctx)
	if err != nil {
		return count, errs.HandlePgError(err)
	}

	return count, nil
}

// GenerateAccessToken generates a new access token for the specified user.
func (s *Store) GenerateAccessToken(userID int32) (Token, error) {
	var token Token

	iat, exp := generateDateClaims(AccessTokenExp)

	value, err := s.generateJWTToken(userID, exp.Unix(), iat.Unix())
	if err != nil {
		return token, err
	}

	token = buildToken(value, iat, exp)

	return token, nil
}

// ParseAndValidateJWT parses and validates a JWT token string, returning its claims if valid.
func (s *Store) ParseAndValidateJWT(tokenString string) (jwt.MapClaims, error) {
	var claims jwt.MapClaims

	secret := s.config.JWTSecret

	token, err := parseJWT(secret, tokenString)
	if err != nil {
		return jwt.MapClaims{}, err
	}

	claims, err = validateJWT(token, s.config)
	if err != nil {
		return claims, err
	}

	return claims, nil
}

// ParseUUID parses a UUID string into a [16]byte array.
func ParseUUID(s string) ([16]byte, error) {
	var out [16]byte

	if len(s) != 36 || s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return out, errs.ErrInvalidUUIDFormat
	}

	var hexBuf [32]byte
	j := 0
	for i := range 36 {
		c := s[i]
		if c == '-' {
			continue
		}
		if !isHex(c) {
			return out, errs.ErrInvalidUUIDHex
		}
		if j >= 32 {
			return out, errs.ErrInvalidUUIDLength
		}
		hexBuf[j] = c
		j++
	}
	if j != 32 {
		return out, errs.ErrInvalidUUIDLength
	}

	if _, err := hex.Decode(out[:], hexBuf[:]); err != nil {
		return out, err
	}

	return out, nil
}

// UUIDToString converts a pgtype.UUID to its string representation.
func UUIDToString(u pgtype.UUID) (string, error) {
	if !u.Valid {
		return "", errs.ErrUUIDIsNull
	}
	b := u.Bytes

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	), nil
}

// generateJWTToken creates and signs a JWT token for the given user ID, expiration, and issued-at times.
func (s *Store) generateJWTToken(userID int32, exp, iat int64) (string, error) {
	if string(s.config.JWTSecret) == "" {
		return "", errs.ErrJWTSecretNotSet
	}

	claims := jwt.MapClaims{
		"sub": userID,
		"iss": s.config.JWTIssuer,
		"aud": s.config.JWTAudience,
		"exp": exp,
		"iat": iat,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.config.JWTSecret)
	if err != nil {
		log.Printf("Could not generate access token to user %d, error: %s", userID, err.Error())

		return "", errs.ErrGenerateJWTToken
	}

	return signedToken, nil
}

// parseJWT parses a JWT token string using the provided secret and returns the token object.
func parseJWT(secret []byte, tokenString string) (*jwt.Token, error) {
	keyFunc := func(_ *jwt.Token) (any, error) {
		return secret, nil
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		jwt.MapClaims{},
		keyFunc,
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// validateJWT checks the validity of a JWT token and extracts its claims.
// Returns an error if the token is invalid or claims cannot be extracted.
func validateJWT(token *jwt.Token, config *app.Config) (jwt.MapClaims, error) {
	if !token.Valid {
		return nil, errs.ErrInvalidJWTToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errs.ErrInvalidClaimsType
	}

	issuer, err := claims.GetIssuer()
	if err != nil {
		return nil, errs.ErrInvalidClaimsType
	}
	if issuer != config.JWTIssuer {
		return nil, errs.ErrInvalidJWTIssuer
	}

	audience, err := claims.GetAudience()
	if err != nil {
		return nil, errs.ErrInvalidClaimsType
	}
	for _, aud := range audience {
		if !slices.Contains(config.JWTAudience, aud) {
			return nil, errs.ErrInvalidJWTAudience
		}
	}

	return claims, nil
}

// generateDateClaims returns the issued-at and expiration times for a token based on the given duration.
func generateDateClaims(expSum time.Duration) (time.Time, time.Time) {
	iat := time.Now().UTC()
	exp := iat.Add(expSum)

	return iat, exp
}

// buildToken constructs a Token struct from the given value, issued-at, and expiration times.
func buildToken(value string, iat, exp time.Time) Token {
	return Token{
		Value:     value,
		IssuedAt:  iat,
		ExpiresAt: exp,
	}
}

// isHex returns true if the given byte is a valid hexadecimal character.
func isHex(c byte) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}
