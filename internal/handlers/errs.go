package handlers

import "errors"

var (
	ErrParseForm  = errors.New("failed to parse form")
	ErrParseField = errors.New("failed to parse field")

	ErrSearchDateFormat  = errors.New("dates must use the YYYY-MM-DD format")
	ErrSearchDateRange   = errors.New("the from date must be on or before the to date")
	ErrSearchTermTooLong = errors.New("search terms must be at most 50 characters")
)
