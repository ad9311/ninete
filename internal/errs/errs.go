// Package errs for errors
package errs

import (
	"errors"
)

// Errors
var (
	ErrEnvNoTSet  = errors.New("env variable not set")
	ErrInvalidEnv = errors.New("invalid ENV")

	ErrCommandExists  = errors.New("command already exists")
	ErrUnknownCommand = errors.New("unknown command")
)
