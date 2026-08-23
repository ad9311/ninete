package logic

import (
	"fmt"
	"strings"
	"unicode"

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

// ValidationError carries the per-field failures alongside the flat message the
// pages already render. The API turns Fields into its 422 body; the templates
// keep printing Error(), which is byte-for-byte what it was before this type
// existed.
type ValidationError struct {
	// Fields maps the snake_case name of each failed field to the rule it
	// broke ("required", "min", "email"). The rule name is the client's to
	// phrase — no message text is invented here.
	Fields  map[string]string
	message string
}

func (e *ValidationError) Error() string { return e.message }

// Unwrap keeps errors.Is(err, ErrValidationFailed) working for every caller
// written before this type.
func (*ValidationError) Unwrap() error { return ErrValidationFailed }

func fmtValidationErrors(err error) error {
	valErr, ok := err.(validator.ValidationErrors)
	if !ok {
		return ErrValidationAssertion
	}

	var chained []string
	fields := make(map[string]string, len(valErr))
	for _, e := range valErr {
		msg := "[" + e.Field() + ":" + e.ActualTag() + "]"
		chained = append(chained, msg)
		fields[snakeFieldName(e.Field())] = e.ActualTag()
	}

	errStr := strings.Join(chained, ",")

	return &ValidationError{
		Fields:  fields,
		message: fmt.Sprintf("%s: %s", ErrValidationFailed, errStr),
	}
}

// snakeFieldName converts a Go field name to the snake_case name the JSON API
// uses, matching the existing form fields (`protein_g`, `saturated_fat_g`) and
// the database columns behind them. A trailing or leading run of capitals stays
// one word, so UserID becomes user_id rather than user_i_d.
func snakeFieldName(field string) string {
	var out []rune

	runes := []rune(field)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			prevIsLower := i > 0 && !unicode.IsUpper(runes[i-1])
			nextIsLower := i+1 < len(runes) && !unicode.IsUpper(runes[i+1])

			if i > 0 && (prevIsLower || nextIsLower) {
				out = append(out, '_')
			}

			out = append(out, unicode.ToLower(r))

			continue
		}

		out = append(out, r)
	}

	return string(out)
}
