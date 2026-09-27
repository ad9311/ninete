package api

import "errors"

var (
	ErrUnauthorized = errors.New(
		"the token was rejected: it is invalid, expired or revoked; create a new one at /account/tokens",
	)
	ErrNotFound        = errors.New("not found")
	ErrTooManyRequests = errors.New("too many failed attempts from this address; wait a minute and retry")
	ErrUnexpected      = errors.New("unexpected response from Ninete")
)
