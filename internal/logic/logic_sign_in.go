package logic

import (
	"context"
	"errors"

	"github.com/ad9311/ninete/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

type SessionParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type NewSession struct {
	User         repo.SafeUser
	RefreshToken Token
	AccessToken  Token
}

func (s *Store) SignInUser(ctx context.Context, params SessionParams) (NewSession, error) {
	var session NewSession

	if err := s.ValidateStruct(params); err != nil {
		return session, err
	}

	user, err := s.FindUserByEmail(ctx, params.Email)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			s.app.Logger.Error("failed to find user by email: %v", err)
		}

		return session, ErrWrongEmailOrPassword
	}

	if err = comparePasswords(params.Password, user.PasswordHash); err != nil {
		return session, err
	}

	refreshToken, err := s.NewRefreshToken(ctx, user.ID)
	if err != nil {
		return session, err
	}

	accessToken, err := s.NewAccessToken(user.ID)
	if err != nil {
		return session, err
	}

	session = NewSession{
		User:         user.ToSafe(),
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}

	return session, nil
}

func comparePasswords(rawPassword, passwordHash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(rawPassword)); err != nil {
		return ErrWrongEmailOrPassword
	}

	return nil
}
