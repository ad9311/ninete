package prog

import "errors"

var (
	ErrEnvNoTSet  = errors.New("environment variable not set")
	ErrInvalidEnv = errors.New("invalid value for ENV")
)
