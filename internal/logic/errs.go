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

	ErrInvalidMood = errors.New("invalid mood selection")

	ErrQuickExpenseFormat      = errors.New("quick expense must be: description, amount, date")
	ErrQuickExpenseDescription = errors.New("description must be between 3 and 50 characters")
	ErrQuickExpenseAmount      = errors.New("invalid amount")
	ErrQuickExpenseDate        = errors.New("invalid date")
)
