package cmd

import "errors"

var (
	ErrCommandExists  = errors.New("command already exists")
	ErrUnknownCommand = errors.New("unknown command")
	ErrMissingArg     = errors.New("missing command argument")
	ErrEmptyArgValue  = errors.New("empty argument value")
)
