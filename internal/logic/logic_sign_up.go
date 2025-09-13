package logic

import (
	"context"
	"errors"

	"github.com/ad9311/ninete/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

type SignUpParams struct {
	Username             string `json:"username"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"passwordConfirmation"`
}

func (s *Store) SignUpUser(ctx context.Context, params SignUpParams) (repo.SafeUser, error) {
	var user repo.User

	if params.Password != params.PasswordConfirmation {
		return user.ToSafe(), ErrUnmatchedPasswords
	}

	// TODO Validate params

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

// hashPassword hashes the provided raw password using bcrypt.
// Returns the hashed password or an error if hashing fails.
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
