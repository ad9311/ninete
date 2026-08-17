package logic

import "errors"

var (
	ErrWithPasswords         = errors.New("failed to save passwords")
	ErrWrongEmailOrPassword  = errors.New("wrong email or password")
	ErrInvalidInvitationCode = errors.New("invalid invitation code")
	ErrInvitationCodeExists  = errors.New("invitation code already exists")
	ErrPasswordConfirmation  = errors.New("password and password confirmation do not match")
	ErrInvitationCodeVerify  = errors.New("failed to verify invitation code")
	ErrLoginLookup           = errors.New("failed to look up account")

	// ErrAccountExists names neither the field that collided nor the value, so
	// a holder of a valid invitation code cannot probe which addresses are
	// already registered.
	ErrAccountExists = errors.New("an account with that username or email already exists")

	// ErrSignUpFailed marks a sign-up that failed for a reason the applicant
	// did nothing to cause. The underlying error is for the log, not the page.
	ErrSignUpFailed = errors.New("failed to create account")

	ErrValidationAssertion = errors.New("failed to assert error type")
	ErrValidationFailed    = errors.New("validation failed")

	ErrTagResolutionFailed = errors.New("failed to resolve tags")

	ErrInvalidMood = errors.New("invalid mood selection")

	ErrQuickExpenseFormat      = errors.New("quick expense must be: description, amount, date[, tags]")
	ErrQuickExpenseDescription = errors.New("description must be between 3 and 50 characters")
	ErrQuickExpenseAmount      = errors.New("invalid amount")
	ErrQuickExpenseDate        = errors.New("invalid date")
	ErrQuickExpenseTags        = errors.New("too many tags, 10 maximum")
	ErrQuickExpenseTagName     = errors.New("each tag must be at most 20 characters")
)
