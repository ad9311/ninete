package errs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validation errors
var (
	ErrValidationAssertion = errors.New("failed asserting validator")
	ErrValidationFailed    = errors.New("validation failed")
)

// FmtValidationErrors formats validation errors from the validator package into a single error string.
// If err is not of type validator.ValidationErrors, it returns ErrValidationAssertion.
func FmtValidationErrors(err error) error {
	valErr, ok := err.(validator.ValidationErrors)
	if !ok {
		return ErrValidationAssertion
	}

	var chained []string
	for _, e := range valErr {
		chain := fmt.Sprintf("[%s]:%s", e.Field(), e.Tag())
		chained = append(chained, chain)
	}

	errStr := strings.Join(chained, ";")
	wrappedErr := fmt.Errorf("%w: %s", ErrValidationFailed, errStr)

	return wrappedErr
}
