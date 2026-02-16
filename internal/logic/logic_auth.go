package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

type SessionParams struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type SignUpParams struct {
	Username             string `validate:"required,alphanumunicode,min=3,max=20"`
	Email                string `validate:"required,email"`
	Password             string `validate:"required,min=8,max=20"`
	PasswordConfirmation string `validate:"required,min=8,max=20"`
	InvitationCode       string `validate:"required"`
}

func (s *Store) Login(ctx context.Context, params SessionParams) (repo.User, error) {
	var user repo.User

	params.Email = prog.NormalizeLowerTrim(params.Email)

	if err := s.ValidateStruct(params); err != nil {
		return user, err
	}

	user, err := s.FindUserForAuth(ctx, params.Email)
	if err != nil {
		return user, ErrWrongEmailOrPassword
	}

	if err = comparePasswords(params.Password, user.PasswordHash); err != nil {
		return user, err
	}

	return user, nil
}

func (s *Store) SignUp(ctx context.Context, params SignUpParams) (User, error) {
	var user User

	params.Username = prog.NormalizeLowerTrim(params.Username)
	params.Email = prog.NormalizeLowerTrim(params.Email)
	params.InvitationCode = prog.NormalizeLowerTrim(params.InvitationCode)

	if err := s.ValidateStruct(params); err != nil {
		return user, err
	}

	if params.Password != params.PasswordConfirmation {
		return user, ErrPasswordConfirmation
	}

	if err := s.ValidateInvitationCode(ctx, params.InvitationCode); err != nil {
		return user, err
	}

	passwordHash, err := HashPassword(params.Password)
	if err != nil {
		return user, err
	}

	user, err = s.CreateUser(ctx, repo.InsertUserParams{
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return user, err
	}

	return user, nil
}

func HashPassword(rawPassword string) ([]byte, error) {
	var passHash []byte

	passHash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return passHash, fmt.Errorf("%w: too long", ErrWithPasswords)
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
