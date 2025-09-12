package serve

import "errors"

// Code messages for error responses
const (
	CodeGeneric   = "Error"
	CodeForbidden = "Forbidden Request"
)

// Predefined errors used for generating standardized error responses in the server.
var (
	ErrContentNotSupported = errors.New("request content not supported")
	ErrOriginNotAllowed    = errors.New("origin not allowed")
	ErrNotPathFound        = errors.New("path not found")
	ErrMethodNotAllowed    = errors.New("method not allowed")
)
