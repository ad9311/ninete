package logic

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrUnmatchedPasswords = errors.New("password do not match")
	ErrPasswordTooLong    = errors.New("password too long")
	ErrUserAlreadyExists  = errors.New("user already exist")

	ErrValidationAssertion = errors.New("failed to assert error type")
	ErrValidationFailed    = errors.New("validation failed")
)

func (s *Store) ValidateStruct(st any) error {
	if err := s.validte.Struct(st); err != nil {
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
		msg := fmt.Sprintf("[%s:%s]", e.Field(), e.ActualTag())
		chained = append(chained, msg)
	}

	errStr := strings.Join(chained, ",")
	wrappedErr := fmt.Errorf("%w: %s", ErrValidationFailed, errStr)

	return wrappedErr
}
