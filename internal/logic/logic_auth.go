package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/ad9311/ninete/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

type SessionParams struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

func (s *Store) Login(ctx context.Context, params SessionParams) (repo.User, error) {
	var user repo.User

	if err := s.ValidateStruct(params); err != nil {
		return user, err
	}

	user, err := s.FindUserByEmail(ctx, params.Email)
	if err != nil {
		return user, ErrWrongEmailOrPassword
	}

	if err = comparePasswords(params.Password, user.PasswordHash); err != nil {
		return user, err
	}

	return user, nil
}

func HashPassword(rawPassword string) ([]byte, error) {
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
