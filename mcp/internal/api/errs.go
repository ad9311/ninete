package api

import "errors"

var (
	ErrUnauthorized = errors.New(
		"the token was rejected: it is invalid, expired or revoked; create a new one at /account/tokens",
	)
	ErrNotFound        = errors.New("not found")
	ErrTooManyRequests = errors.New("too many failed attempts from this address; wait a minute and retry")
	ErrUnexpected      = errors.New("unexpected response from Ninete")

	// ErrQueryTag is a malformed `query:"…"` tag on a query struct: a
	// programming error, which contracttest surfaces on every struct.
	ErrQueryTag = errors.New("invalid query tag")
	// ErrPathQuery is a request path carrying its own query string or
	// fragment, which only query structs may supply.
	ErrPathQuery = errors.New("a request path may not carry a query string; pass a query struct to Get")
	// ErrQueryValue is a value a query struct's own rule refuses, caught
	// before the request is sent.
	ErrQueryValue = errors.New("invalid query value")
)
