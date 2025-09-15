package logic

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrUnmatchedPasswords   = errors.New("passwords do not match")
	ErrPasswordTooLong      = errors.New("password too long")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrWrongEmailOrPassword = errors.New("wrong email or password")

	ErrValidationAssertion = errors.New("failed to assert error type")
	ErrValidationFailed    = errors.New("validation failed")

	ErrNotFound = errors.New("resource not found")

	ErrInvalidJWTToken = errors.New("invalid jwt token")
)

func (s *Store) ValidateStruct(st any) error {
	if err := s.validate.Struct(st); err != nil {
		return fmtValidationErrors(err)
	}

	return nil
}

func fmtValidationErrors(err error) error {
	valErr, ok := err.(validator.ValidationErrors)
	if !ok {
		return ErrValidationAssertion
	}

	var chained []string
	for _, e := range valErr {
		msg := "[" + e.Field() + ":" + e.ActualTag() + "]"
		chained = append(chained, msg)
	}

	errStr := strings.Join(chained, ",")
	wrappedErr := fmt.Errorf("%w: %s", ErrValidationFailed, errStr)

	return wrappedErr
}

func HandleDBError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	return err
}
