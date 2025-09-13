package serve

import "errors"

// Code messages for error responses
const (
	CodeGeneric   = "Error"
	CodeForbidden = "Forbidden Request"
	CodeBadFormat = "Bad Format"
)

// Predefined errors used for generating standardized error responses in the server.
var (
	ErrContentNotSupported = errors.New("request content not supported")
	ErrOriginNotAllowed    = errors.New("origin not allowed")
	ErrNotPathFound        = errors.New("path not found")
	ErrMethodNotAllowed    = errors.New("method not allowed")
	ErrFormParsing         = errors.New("failed to parse form")
)
