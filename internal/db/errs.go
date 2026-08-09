package db

import "errors"

var (
	ErrForeignKeysDisabled = errors.New("foreign key enforcement is off on a new connection")
	ErrPragmaNoValue       = errors.New("pragma returned no value")
)
