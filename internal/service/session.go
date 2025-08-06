package service

import (
	"context"
	"errors"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// SessionParams contains the parameters required for signing in.
type SessionParams struct {
	Email    string `validate:"required,email" json:"email"` // User's email address
	Password string `validate:"required" json:"password"`    // User's password
}

// SafeUser contains user information excluding the password hash.
type SafeUser struct {
	ID       int32  `json:"id"`       // User ID
	Username string `json:"username"` // Username
	Email    string `json:"email"`    // Email address
}

// SessionObject represents the session data returned when a user signs in successfully.
type SessionObject struct {
	User         SafeUser // Authenticated user information
	AccessToken  Token    // JWT access token
	RefreshToken Token    // Refresh token
}

// SignInUser validates credentials, authenticates the user, and returns a session object with tokens.
func (s *Store) SignInUser(ctx context.Context, params SessionParams) (SessionObject, error) {
	var sessionObject SessionObject

	if err := s.validate.Struct(params); err != nil {
		err := errs.FmtValidationErrors(err)

		return sessionObject, err
	}

	user, err := s.FindUserByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return sessionObject, errs.ErrWrongEmailOrPassword
		}

		return sessionObject, err
	}

	if err = comparePasswords(params.Password, user.PasswordHash); err != nil {
		return sessionObject, err
	}

	accessToken, err := s.GenerateAccessToken(user.ID)
	if err != nil {
		return sessionObject, err
	}

	refreshToken, err := s.SaveRefreshToken(ctx, user.ID)
	if err != nil {
		return sessionObject, err
	}

	uuidStr, err := UUIDToString(refreshToken.Uuid)
	if err != nil {
		return sessionObject, err
	}

	sessionObject = SessionObject{
		User: SafeUser{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
		AccessToken: accessToken,
		RefreshToken: Token{
			Value:     uuidStr,
			IssuedAt:  refreshToken.IssuedAt.Time,
			ExpiresAt: refreshToken.ExpiresAt.Time,
		},
	}

	return sessionObject, nil
}

// SignOutUser deletes the user's refresh token from the database, effectively signing them out.
func (s *Store) SignOutUser(ctx context.Context, uuidString string) error {
	uuid, err := ParseUUID(uuidString)
	if err != nil {
		return err
	}

	pgUUID := pgtype.UUID{
		Bytes: uuid,
		Valid: true,
	}

	if err := s.queries.DeleteRefreshTokenWhereUUID(ctx, pgUUID); err != nil {
		return errs.HandlePgError(err)
	}

	return nil
}

// comparePasswords compares a raw password with its hashed value and returns an error if they do not match.
func comparePasswords(rawPassword, passwordHash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(rawPassword)); err != nil {
		return errs.ErrWrongEmailOrPassword
	}

	return nil
}
