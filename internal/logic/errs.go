package logic

import "errors"

var (
	ErrUnmatchedPasswords = errors.New("password do not match")
	ErrPasswordTooLong    = errors.New("password too long")
)
