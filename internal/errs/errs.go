// Package errs provides common error variables used throughout the application.
package errs

import (
	"errors"
)

// Global application errors
var (
	ErrEnvNoTSet  = errors.New("env variable not set")
	ErrInvalidEnv = errors.New("invalid value for ENV")

	ErrCommandExists  = errors.New("command already exists")
	ErrUnknownCommand = errors.New("unknown command")
)
