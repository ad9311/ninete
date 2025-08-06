package errs

import "errors"

// Command/CLI errors
var (
	ErrWithCommand              = errors.New("error with command")
	ErrCommandAlreadyRegistered = errors.New("command already registered")
	ErrUnknowCommand            = errors.New("unknown command")
	ErrServiceFuncNotAvailable  = errors.New("service not available for this environment")
)
