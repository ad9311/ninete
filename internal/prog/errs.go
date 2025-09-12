package prog

import "errors"

// Global program errors
var (
	ErrEnvNoTSet       = errors.New("env variable not set")
	ErrInvalidEnv      = errors.New("invalid value for ENV")
	ErrInterfaceNotSet = errors.New("interface not set")
)
