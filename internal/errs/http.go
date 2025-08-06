package errs

import (
	"errors"
)

// HTTP/server errors
var (
	ErrFormParsing               = errors.New("error parsing the form")
	ErrMethodNotAllowedForOrigin = errors.New("method not allowed for this origin")
	ErrMethodNotAllowed          = errors.New("method not allowed for this endpoint")
	ErrUnsupportedMediaType      = errors.New("Content-Type must be application/json")
	ErrNotFound                  = errors.New("resource not found")
	ErrStandard                  = errors.New("error")
)

// UnknownError wraps the standard error with a custom message.
func UnknownError(msg string) error {
	return WrapMessageWithError(ErrStandard, msg)
}
