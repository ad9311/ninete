package logic

import "errors"

var (
	ErrWithPasswords        = errors.New("failed to save passwords")
	ErrWrongEmailOrPassword = errors.New("wrong email or password")

	ErrValidationAssertion = errors.New("failed to assert error type")
	ErrValidationFailed    = errors.New("validation failed")
)
