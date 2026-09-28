package api

import (
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// The query strings this module sends are declared once, as `query:"…"` tags
// on the structs below, in the same grammar the app uses for the keys it reads
// (internal/handlers/query.go):
//
//	query:"page,positive"              a whole number of at least 1
//	query:"start,integer,required"     a whole number, always sent
//	query:"archived,boolean"           true or false
//	query:"mode,oneof=month months"    one of the listed values
//	query:"q,max=50"                   at most 50 characters
//
// contracttest compares QueryRules of each struct with contract/api.json, so a
// key or rule that disagrees with the server fails a test without any tool
// having to send it. A oneof here may list fewer values than the server
// accepts — a tool may offer a subset — but never one the server does not.
// The tools read their allowlists back from these tags (QueryOneOf), so the
// list a tool validates its input against is the list this check covers.
//
// EncodeQuery is the only way the tools build a query string; forbidigo keeps
// url.Values out of internal/tools.

// DashboardQuery is GET /api/dashboard.
type DashboardQuery struct {
	ThisStart int64 `query:"this_start,integer,required"`
	ThisEnd   int64 `query:"this_end,integer,required"`
	LastStart int64 `query:"last_start,integer,required"`
	LastEnd   int64 `query:"last_end,integer,required"`
}

// ExpenseStatsQuery is GET /api/expenses/stats. The tool sorts nothing, so it
// declares no sort pair.
type ExpenseStatsQuery struct {
	Start int64 `query:"start,integer"`
	End   int64 `query:"end,integer"`
}

// BudgetsQuery is GET /api/expenses/budgets.
type BudgetsQuery struct {
	Start int64  `query:"start,integer,required"`
	End   int64  `query:"end,integer,required"`
	Mode  string `query:"mode,oneof=month months"`
}

// ExpenseListQuery is GET /api/expenses. Its sort_field lists the columns
// search_expenses offers, a subset of what the server accepts.
type ExpenseListQuery struct {
	Q            string `query:"q,max=50"`
	Tag          string `query:"tag,max=50"`
	CategoryID   int    `query:"category_id,positive"`
	Start        int64  `query:"start,integer"`
	End          int64  `query:"end,integer"`
	CreatedStart int64  `query:"created_start,integer"`
	CreatedEnd   int64  `query:"created_end,integer"`
	Page         int    `query:"page,positive"`
	PerPage      int    `query:"per_page,oneof=15 25 50 100"`
	SortField    string `query:"sort_field,oneof=created_at date amount description"`
	SortOrder    string `query:"sort_order,oneof=ASC DESC"`
}

// RecurrentExpenseListQuery is GET /api/recurrent-expenses.
type RecurrentExpenseListQuery struct {
	Archived   bool   `query:"archived,boolean"`
	CategoryID int    `query:"category_id,positive"`
	Page       int    `query:"page,positive"`
	PerPage    int    `query:"per_page,oneof=15 25 50 100"`
	SortField  string `query:"sort_field,oneof=created_at description amount period"`
	SortOrder  string `query:"sort_order,oneof=ASC DESC"`
}

// queryRule is one parsed `query:"…"` tag.
type queryRule struct {
	key      string
	kind     string // "string", "integer", "positive", "boolean" or "oneof"
	oneOf    []string
	max      int
	required bool
}

func parseQueryRule(tag string) (queryRule, error) {
	parts := strings.Split(tag, ",")
	rule := queryRule{key: parts[0], kind: "string"}

	for _, part := range parts[1:] {
		name, arg, _ := strings.Cut(part, "=")

		switch name {
		case "integer", "positive", "boolean":
			rule.kind = name
		case "oneof":
			rule.kind = name
			rule.oneOf = strings.Fields(arg)
		case "max":
			n, err := strconv.Atoi(arg)
			if err != nil || n < 1 {
				return rule, fmt.Errorf("%w %q: max needs a positive number", ErrQueryTag, tag)
			}

			rule.max = n
		case "required":
			rule.required = true
		default:
			return rule, fmt.Errorf("%w %q: unknown rule %q", ErrQueryTag, tag, name)
		}
	}

	return rule, nil
}

// describe matches queryRule.describe in internal/handlers/query.go, which
// writes contract/api.json.
func (rule queryRule) describe() string {
	desc := rule.kind
	switch {
	case rule.kind == "positive":
		desc = "positive integer"
	case rule.kind == "oneof":
		desc = "oneof<" + strings.Join(rule.oneOf, "|") + ">"
	case rule.max > 0:
		desc += "<max " + strconv.Itoa(rule.max) + ">"
	}

	if rule.required {
		desc = "required " + desc
	}

	return desc
}

// accepts mirrors the server's check on a formatted, non-empty value.
func (rule queryRule) accepts(value string) bool {
	switch rule.kind {
	case "integer":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return false
		}
	case "positive":
		if n, err := strconv.ParseInt(value, 10, 64); err != nil || n < 1 {
			return false
		}
	case "boolean":
		if value != "true" && value != "false" {
			return false
		}
	case "oneof":
		if !slices.Contains(rule.oneOf, value) {
			return false
		}
	}

	return rule.max == 0 || utf8.RuneCountInString(value) <= rule.max
}

// fits reports whether a Go field kind can carry the rule's values.
func (rule queryRule) fits(kind reflect.Kind) bool {
	switch rule.kind {
	case "integer", "positive":
		return kind == reflect.Int || kind == reflect.Int64
	case "boolean":
		return kind == reflect.Bool
	case "oneof":
		return kind == reflect.String || kind == reflect.Int
	default:
		return kind == reflect.String
	}
}

type queryField struct {
	rule  queryRule
	value reflect.Value
}

func queryFields(v any) ([]queryField, error) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: %T is not a struct", ErrQueryTag, v)
	}

	var fields []queryField

	for i := range rv.NumField() {
		field := rv.Type().Field(i)

		tag := field.Tag.Get("query")
		if tag == "" {
			continue
		}

		rule, err := parseQueryRule(tag)
		if err != nil {
			return nil, err
		}

		if !rule.fits(field.Type.Kind()) {
			return nil, fmt.Errorf("%w %q: a %s field cannot carry it", ErrQueryTag, tag, field.Type.Kind())
		}

		fields = append(fields, queryField{rule, rv.Field(i)})
	}

	return fields, nil
}

// EncodeQuery turns a query struct into the values to send. A zero field is
// left out unless it is required. A value its own rule refuses is an error
// rather than a request: the server would answer it with a 422 anyway, and
// failing here names the key before anything is sent.
func EncodeQuery(v any) (url.Values, error) {
	fields, err := queryFields(v)
	if err != nil {
		return nil, err
	}

	query := url.Values{}

	for _, f := range fields {
		if f.value.IsZero() && !f.rule.required {
			continue
		}

		var value string

		switch f.value.Kind() {
		case reflect.Int, reflect.Int64:
			value = strconv.FormatInt(f.value.Int(), 10)
		case reflect.Bool:
			value = strconv.FormatBool(f.value.Bool())
		default:
			value = f.value.String()
		}

		if !f.rule.accepts(value) {
			return nil, fmt.Errorf("%w: %s=%q, the server expects %s", ErrQueryValue, f.rule.key, value, f.rule.describe())
		}

		query.Set(f.rule.key, value)
	}

	return query, nil
}

// QueryRules maps each key of a query struct to its rule, described the way
// contract/api.json records the server's.
func QueryRules(v any) (map[string]string, error) {
	fields, err := queryFields(v)
	if err != nil {
		return nil, err
	}

	rules := make(map[string]string, len(fields))
	for _, f := range fields {
		rules[f.rule.key] = f.rule.describe()
	}

	return rules, nil
}

// QueryOneOf returns the values a oneof key accepts, in declared order, for a
// tool to validate its input against and to name in its error message. It
// panics on a key that is not a oneof: the arguments are constants, and every
// tool that calls this runs in the tools test.
func QueryOneOf(v any, key string) []string {
	fields, err := queryFields(v)
	if err != nil {
		panic(err)
	}

	for _, f := range fields {
		if f.rule.key == key && f.rule.kind == "oneof" {
			return f.rule.oneOf
		}
	}

	panic(fmt.Sprintf("%T has no oneof key %q", v, key))
}
