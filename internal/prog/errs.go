package prog

import "errors"

var (
	ErrEnvNoTSet  = errors.New("environment variable not set")
	ErrInvalidEnv = errors.New("invalid value for ENV")
	ErrParsing    = errors.New("failed to parse value")
)
