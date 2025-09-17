package logic

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ExpRefreshToken = 24 * 7 * time.Hour
	ExpAccessToken  = 15 * time.Minute
)

type Token struct {
	Value     string `json:"value"`
	IssuedAt  int64  `json:"issuedAt"`
	ExpiresAt int64  `json:"expiresAt"`
}

func (s *Store) NewRefreshToken(ctx context.Context, userID int) (Token, error) {
	var token Token

	value, err := randomRefreshToken()
	if err != nil {
		return token, err
	}

	iat, exp := generateDateClaims(ExpRefreshToken)

	_, err = s.queries.InsertRefreshToken(ctx, repo.InsertRefreshTokenParams{
		UserID:    userID,
		TokenHash: hashToken(value),
		IssuedAt:  iat,
		ExpiresAt: exp,
	})
	if err != nil {
		return token, HandleDBError(err)
	}

	token = Token{
		Value:     value,
		IssuedAt:  iat,
		ExpiresAt: exp,
	}

	return token, nil
}

func (s *Store) NewAccessToken(userID int) (Token, error) {
	iat, exp := generateDateClaims(ExpAccessToken)

	claims := jwt.MapClaims{
		"sub": strconv.Itoa(userID),
		"iss": s.tokenVars.jwtIssuer,
		"aud": s.tokenVars.jwtAudience,
		"exp": exp,
		"iat": iat,
	}

	var token Token
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	value, err := jwtToken.SignedString([]byte(s.tokenVars.jwtSecret))
	if err != nil {
		s.app.Logger.Errorf("Failed to generate access token for user %d: %v", userID, err)

		return token, fmt.Errorf("failed to generate access token: %w", err)
	}

	token = Token{
		Value:     value,
		IssuedAt:  iat,
		ExpiresAt: exp,
	}

	return token, nil
}

func (s *Store) ParseAndValidateJWT(tokenString string) (jwt.MapClaims, error) {
	var claims jwt.MapClaims

	token, err := s.parseJWT(tokenString)
	if err != nil {
		return jwt.MapClaims{}, err
	}

	claims, err = s.validateJWT(token)
	if err != nil {
		return claims, err
	}

	return claims, nil
}

func (s *Store) parseJWT(tokenString string) (*jwt.Token, error) {
	keyFunc := func(_ *jwt.Token) (any, error) {
		return []byte(s.tokenVars.jwtSecret), nil
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

func (s *Store) validateJWT(token *jwt.Token) (jwt.MapClaims, error) {
	if !token.Valid {
		return nil, ErrInvalidJWTToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("%w, invalid claims", ErrInvalidJWTToken)
	}

	issuer, err := claims.GetIssuer()
	if err != nil {
		return nil, fmt.Errorf("%w, invalid issuer", ErrInvalidJWTToken)
	}
	if issuer != s.tokenVars.jwtIssuer {
		return nil, fmt.Errorf("%w, invalid issuer", ErrInvalidJWTToken)
	}

	audience, err := claims.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("%w, invalid audience", ErrInvalidJWTToken)
	}
	hasAudience := false
	for _, aud := range audience {
		if slices.Contains(s.tokenVars.jwtAudience, aud) {
			hasAudience = true

			break
		}
	}
	if !hasAudience {
		return nil, fmt.Errorf("%w, invalid audience", ErrInvalidJWTToken)
	}

	return claims, nil
}

func generateDateClaims(dur time.Duration) (int64, int64) {
	iat := time.Now().UTC()
	exp := iat.Add(dur).UTC()

	return iat.Unix(), exp.Unix()
}

func randomRefreshToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("failed to generate random refresh token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))

	return sum[:]
}
