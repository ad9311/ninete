package cmd

import "errors"

// Command specific errors
var (
	ErrCommandExists  = errors.New("command already exists")
	ErrUnknownCommand = errors.New("unknown command")
)
