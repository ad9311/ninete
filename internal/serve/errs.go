package serve

import "errors"

var (
	ErrContentNotSupported = errors.New("request content not supported")
	ErrOriginNotAllowed    = errors.New("origin not allowed")
	ErrNotPathFound        = errors.New("path not found")
	ErrMethodNotAllowed    = errors.New("method not allowed")
	ErrFormParsing         = errors.New("failed to parse form")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrInvalidAuthCreds    = errors.New("invalid auth credentials")
)
