package service

import (
	"context"
	"errors"
	"log"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

// RegistrationParams contains the parameters required for signing up a new user.
type RegistrationParams struct {
	Username             string `validate:"required,username,min=3,max=20" json:"username"`    // Desired username
	Email                string `validate:"required,email" json:"email"`                       // User's email address
	Password             string `validate:"required,min=8,max=20" json:"password"`             // User's password
	PasswordConfirmation string `validate:"required,min=8,max=20" json:"passwordConfirmation"` // Confirmation of user's password
}

// RegisterUser validates registration parameters and saves the new user in the database.
// Returns the created user or an error if validation or insertion fails.
func (s *Store) RegisterUser(ctx context.Context, params RegistrationParams) (repo.User, error) {
	var user repo.User

	if params.Password != params.PasswordConfirmation {
		return user, errs.ErrUnmatchedPasswords
	}

	if err := s.validate.Struct(params); err != nil {
		err := errs.FmtValidationErrors(err)

		return user, err
	}

	hashedPassword, err := hashPassword(params.Password)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return user, errs.ErrPasswordTooLong
		}

		log.Printf("Failed to has password for %s, error: %s", params.Username, err.Error())

		return user, errs.ErrHashingPassword
	}

	insertParams := repo.InsertUserParams{
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: string(hashedPassword),
	}

	user, err = s.queries.InsertUser(ctx, insertParams)
	if err != nil {
		return user, errs.HandlePgError(err)
	}

	return user, nil
}

// hashPassword hashes the provided raw password using bcrypt.
// Returns the hashed password or an error if hashing fails.
func hashPassword(rawPassword string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return []byte(""), err
	}

	return hashedPassword, nil
}
