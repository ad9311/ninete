package logic

import (
	"context"
	"errors"

	"github.com/ad9311/ninete/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

type SignUpParams struct {
	Username             string `json:"username" validate:"required,min=3,max=20"` // TODO validate format
	Email                string `json:"email" validate:"email"`
	Password             string `json:"password" validate:"min=8,max=20"` // TODO validate format
	PasswordConfirmation string `json:"passwordConfirmation" validate:"min=8,max=20"`
}

func (s *Store) SignUpUser(ctx context.Context, params SignUpParams) (repo.SafeUser, error) {
	var user repo.User

	if params.Password != params.PasswordConfirmation {
		return user.ToSafe(), ErrUnmatchedPasswords
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
		return user.ToSafe(), err
	}

	return user.ToSafe(), nil
}

func hashPassword(rawPassword string) ([]byte, error) {
	var passHash []byte

	passHash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return passHash, ErrPasswordTooLong
		}

		return passHash, err
	}

	return passHash, nil
}
