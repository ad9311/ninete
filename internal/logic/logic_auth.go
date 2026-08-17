package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

// dummyHash gives Login something to compare against when the email matches no
// account, so both failure paths pay the same bcrypt cost. The plaintext behind
// it is irrelevant and never matches a real account, since no user is created
// with it. Hashing is deferred to the first failed login rather than paid at
// startup.
var dummyHash = sync.OnceValue(func() []byte { //nolint:gochecknoglobals // one-time constant
	// GenerateFromPassword only fails on an input over 72 bytes, and this one is
	// a fixed short literal.
	hash, _ := bcrypt.GenerateFromPassword([]byte("ninete-no-such-account"), bcrypt.DefaultCost)

	return hash
})

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
		// A lookup that failed for any reason other than "no such row" is a
		// server fault, not a credential problem. Reporting it as one hides
		// real database failures behind a login form error.
		if !errors.Is(err, sql.ErrNoRows) {
			return user, fmt.Errorf("%w: %w", ErrLoginLookup, err)
		}

		// Spend the same bcrypt time a real comparison would, so response
		// latency does not reveal whether the email owns an account.
		_ = comparePasswords(params.Password, dummyHash())

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

func comparePasswords(rawPassword string, passwordHash []byte) error {
	if err := bcrypt.CompareHashAndPassword(passwordHash, []byte(rawPassword)); err != nil {
		return ErrWrongEmailOrPassword
	}

	return nil
}
