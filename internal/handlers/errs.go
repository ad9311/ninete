package handlers

import "errors"

var (
	ErrParseForm  = errors.New("failed to parse form")
	ErrParseField = errors.New("failed to parse field")

	// ErrLoginUnavailable is what the browser sees when the account lookup
	// itself failed. The underlying error goes to the log instead, so a
	// database fault does not describe itself on a public page.
	ErrLoginUnavailable = errors.New("login is temporarily unavailable, please try again")

	ErrSearchDateFormat  = errors.New("dates must use the YYYY-MM-DD format")
	ErrSearchDateRange   = errors.New("the from date must be on or before the to date")
	ErrSearchTermTooLong = errors.New("search terms must be at most 50 characters")
)
