// Package errs for errors
package errs

import (
	"errors"
)

// Errors
var (
	ErrEnvNoTSet  = errors.New("env variable not set")
	ErrInvalidEnv = errors.New("invalid ENV")
)
