package serve

import "errors"

const (
	CodeGeneric   = "Error"
	CodeForbidden = "Forbidden Request"
	CodeBadFormat = "Bad Format"
)

var (
	ErrContentNotSupported = errors.New("request content not supported")
	ErrOriginNotAllowed    = errors.New("origin not allowed")
	ErrNotPathFound        = errors.New("path not found")
	ErrMethodNotAllowed    = errors.New("method not allowed")
	ErrFormParsing         = errors.New("failed to parse form")
)
