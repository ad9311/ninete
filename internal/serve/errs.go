package serve

import "errors"

var (
	ErrNotAllowed     = errors.New("request not allowed")
	ErrLayoutNotFound = errors.New("layout template not found")
)
