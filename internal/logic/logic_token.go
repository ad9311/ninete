package logic

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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
	IssuedAt  int64  `json:"issued_at"`
	ExpiresAt int64  `json:"expires_at"`
}

func (s *Store) NewRefreshToken(ctx context.Context, userID int) (Token, error) {
	var token Token

	value, err := RandomRefreshToken()
	if err != nil {
		return token, err
	}

	iat, exp := generateDateClaims(ExpRefreshToken)

	_, err = s.queries.InsertRefreshToken(ctx, repo.InsertRefreshTokenParams{
		UserID:    userID,
		TokenHash: HashToken(value),
		IssuedAt:  iat,
		ExpiresAt: exp,
	})
	if err != nil {
		return token, err
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

	value, err := jwtToken.SignedString(s.tokenVars.jwtSecret)
	if err != nil {
		s.app.Logger.Error("Failed to generate access token for user %d: %v", userID, err)

		return token, nil
	}

	token = Token{
		Value:     value,
		IssuedAt:  iat,
		ExpiresAt: exp,
	}

	return token, nil
}

func generateDateClaims(dur time.Duration) (int64, int64) {
	iat := time.Now().UTC()
	exp := iat.Add(dur).UTC()

	return iat.Unix(), exp.Unix()
}

func RandomRefreshToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("failed to generate random refresh token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))

	return sum[:]
}
