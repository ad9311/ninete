package cmd

import "errors"

var (
	ErrCommandExists  = errors.New("command already exists")
	ErrUnknownCommand = errors.New("unknown command")
)
