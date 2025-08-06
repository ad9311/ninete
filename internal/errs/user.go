package errs

import "errors"

// User/session/registration errors
var (
	ErrUnmatchedPasswords   = errors.New("passwords do not match")
	ErrHashingPassword      = errors.New("could not hash user password")
	ErrPasswordTooLong      = errors.New("password is too long")
	ErrWrongEmailOrPassword = errors.New("incorrect email or password")
	ErrUserHasRole          = errors.New("user already has role")
	ErrEmptyRoleName        = errors.New("empty role name")
)
