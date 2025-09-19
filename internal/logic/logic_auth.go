package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/ad9311/ninete/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

type SignUpParams struct {
	Username             string `json:"username" validate:"required,min=3,max=20"` // TODO validate format
	Email                string `json:"email" validate:"email"`
	Password             string `json:"password" validate:"min=8,max=20"` // TODO validate format
	PasswordConfirmation string `json:"passwordConfirmation" validate:"min=8,max=20"`
}

type SessionParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type NewSession struct {
	User         repo.SafeUser
	RefreshToken Token
	AccessToken  Token
}

func (s *Store) SignUpUser(ctx context.Context, params SignUpParams) (repo.SafeUser, error) {
	var user repo.User

	if params.Password != params.PasswordConfirmation {
		return user.ToSafe(), fmt.Errorf("%w, they do not match", ErrWithPasswords)
	}

	if err := s.ValidateStruct(params); err != nil {
		return user.ToSafe(), err
	}

	passwordHash, err := hashPassword(params.Password)
	if err != nil {
		return user.ToSafe(), err
	}

	user, err = s.queries.InsertUser(ctx, repo.InsertUserParams{
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return user.ToSafe(), HandleDBError(err)
	}

	return user.ToSafe(), nil
}

func (s *Store) SignInUser(ctx context.Context, params SessionParams) (NewSession, error) {
	var session NewSession

	if err := s.ValidateStruct(params); err != nil {
		return session, err
	}

	user, err := s.FindUserByEmail(ctx, params.Email)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			s.app.Logger.Errorf("failed to find user by email: %v", err)
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

func (s *Store) SignOutUser(ctx context.Context, tokenStr string) error {
	tokenHash := hashToken(tokenStr)

	_, err := s.queries.DeleteRefreshToken(ctx, tokenHash)
	if err != nil {
		return HandleDBError(err)
	}

	return nil
}

func hashPassword(rawPassword string) ([]byte, error) {
	var passHash []byte

	passHash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return passHash, fmt.Errorf("%w, too long", ErrWithPasswords)
		}

		return passHash, err
	}

	return passHash, nil
}

func comparePasswords(rawPassword, passwordHash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(rawPassword)); err != nil {
		return ErrWrongEmailOrPassword
	}

	return nil
}
