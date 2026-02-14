package logic

import "errors"

var (
	ErrWithPasswords         = errors.New("failed to save passwords")
	ErrWrongEmailOrPassword  = errors.New("wrong email or password")
	ErrInvalidInvitationCode = errors.New("invalid invitation code")
	ErrInvitationCodeExists  = errors.New("invitation code already exists")
	ErrPasswordConfirmation  = errors.New("password and password confirmation do not match")
	ErrInvitationCodeVerify  = errors.New("failed to verify invitation code")

	ErrValidationAssertion = errors.New("failed to assert error type")
	ErrValidationFailed    = errors.New("validation failed")

	ErrTagResolutionFailed = errors.New("failed to resolve tags")
)
