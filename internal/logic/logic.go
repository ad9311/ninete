package logic

import (
	"fmt"
	"strings"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-playground/validator/v10"
)

type Store struct {
	app      *prog.App
	queries  repo.Queries
	validate *validator.Validate
}

func New(app *prog.App, queries repo.Queries) *Store {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &Store{
		app:      app,
		queries:  queries,
		validate: validate,
	}
}

func (s *Store) ValidateStruct(st any) error {
	if err := s.validate.Struct(st); err != nil {
		return fmtValidationErrors(err)
	}

	return nil
}

func fmtValidationErrors(err error) error {
	valErr, ok := err.(validator.ValidationErrors)
	if !ok {
		return ErrValidationAssertion
	}

	var chained []string
	for _, e := range valErr {
		msg := "[" + e.Field() + ":" + e.ActualTag() + "]"
		chained = append(chained, msg)
	}

	errStr := strings.Join(chained, ",")
	wrappedErr := fmt.Errorf("%w: %s", ErrValidationFailed, errStr)

	return wrappedErr
}
