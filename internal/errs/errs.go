// Package errs provides utility functions for error wrapping and formatting.
package errs

import "fmt"

// WrapErrorWithMessage wraps an existing error with a custom message.
func WrapErrorWithMessage(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}

// WrapMessageWithError wraps a message after an existing error.
func WrapMessageWithError(err error, msg string) error {
	return fmt.Errorf("%w: %s", err, msg)
}
