package logic

import (
	"errors"
	"fmt"
	"sort"
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

// underField re-keys every entry of a ValidationError under a single parent
// field name, for a use-case that validates a *secondary* params struct on its
// way to satisfying the request.
//
// ValidationError.Fields is keyed by validator.FieldError.Field(), the leaf Go
// field name, which says nothing about which struct the field belongs to. A
// nested failure therefore lands in the same flat map as the request's own
// fields, under a key the request payload never carried: a too-long tag on
// POST /api/expenses reported {"name": "max"}, naming a field that endpoint
// does not have, while the offending value arrived under "tags". A client
// highlighting inputs by key marks nothing, and on an endpoint that does own a
// "name" field it would mark the wrong input — or silently overwrite the
// parent's entry for the same name.
//
// The rule each nested field broke is kept; only the key changes. Callers name
// the parent field because only the use-case knows which request key the
// nested struct was built from.
func underField(err error, parent string) error {
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		return err
	}

	fields := make(map[string]string, len(valErr.Fields))
	chained := make([]string, 0, len(valErr.Fields))

	for _, rule := range sortedRules(valErr.Fields) {
		fields[parent] = rule
		chained = append(chained, "["+parent+":"+rule+"]")
	}

	return &ValidationError{
		Fields:  fields,
		message: fmt.Sprintf("%s: %s", ErrValidationFailed, strings.Join(chained, ",")),
	}
}

// sortedRules returns the broken rules in a stable order, so the message a
// nested failure produces does not depend on Go's map iteration.
func sortedRules(fields map[string]string) []string {
	rules := make([]string, 0, len(fields))
	for _, rule := range fields {
		rules = append(rules, rule)
	}

	sort.Strings(rules)

	return rules
}

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
