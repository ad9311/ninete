package handlers

import "errors"

var (
	ErrParseForm  = errors.New("failed to parse form")
	ErrParseField = errors.New("failed to parse field")
)
