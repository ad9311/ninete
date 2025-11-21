package serve

import "errors"

var (
	ErrNotAllowed       = errors.New("request not allowed")
	ErrFormParsing      = errors.New("failed to parse form")
	ErrInvalidAuthCreds = errors.New("invalid auth credentials")
	ErrMissingContext   = errors.New("missing context")
	ErrInvalidId        = errors.New("invalid id")
)
