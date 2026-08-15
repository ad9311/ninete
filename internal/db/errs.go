package db

import "errors"

var (
	ErrForeignKeysDisabled = errors.New("foreign key enforcement is off on a new connection")
	ErrPragmaNoValue       = errors.New("pragma returned no value")
	ErrUnknownEnvStamp     = errors.New("no environment stamp defined for ENV")
	ErrEnvStampRead        = errors.New("failed to read the database environment stamp")
	ErrEnvStampWrite       = errors.New("failed to write the database environment stamp")
	ErrEnvStampMismatch    = errors.New("database environment stamp does not match ENV")
)
