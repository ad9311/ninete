package handlers

import (
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ad9311/ninete/internal/logic"
)

// This file is the only place in the package that reads a request's query
// string. Every other file declares the keys it reads as `query:"…"` tags on a
// struct and fills it with decodeQuery, so the tags are the single declaration
// of a route's keys and of the values each accepts: apiContract lists the
// struct, TestAPIContract records every key's rule in contract/api.json, and a
// renamed key or a changed rule changes that file by itself. forbidigo in
// .golangci.yml rejects any other query read in the package, since a stray
// q.Get("…") literal would bypass the contract.
//
// A tag is the key followed by its rules, comma-separated:
//
//	query:"page,positive"              a whole number of at least 1
//	query:"start,integer,required"     a whole number, and the key must be sent
//	query:"archived,boolean"           exactly "true" or "false"
//	query:"mode,oneof=month months"    one of the listed values, exactly
//	query:"q,max=50"                   a string of at most 50 characters
//
// A key without a type rule is a free string. An empty value counts as absent,
// and an absent key is always accepted unless it is required: the handler
// applies its own default. A present value that breaks its rule is a 422 —
// never a silent fallback, because a client sending a value the server does not
// know (a renamed mode, a page size the UI no longer offers) must find out
// rather than get the default back. The MCP server in mcp/ declares its side
// with the same grammar (mcp/internal/api/query.go).
//
// Fields are strings: decoding validates and copies, and each handler converts
// what it needs. The mcp module's contracttest compares its descriptions with
// these, so the two parsers are kept honest by the same file.

// apiBoundsQuery is the optional [start, end) billed-date pair the
// /api/expenses* routes take (§3.6 of docs/spa-migration.md).
type apiBoundsQuery struct {
	Start string `query:"start,integer"`
	End   string `query:"end,integer"`
}

// apiListQuery is the pagination and category filter every listing reads
// through userScopedQueryOpts. The sort pair lives on each route's own struct,
// because the fields a listing may be sorted by differ.
type apiListQuery struct {
	Page       string `query:"page,positive"`
	PerPage    string `query:"per_page,oneof=15 25 50 100"`
	CategoryID string `query:"category_id,positive"`
}

// apiExpenseListQuery is GET /api/expenses.
type apiExpenseListQuery struct {
	apiBoundsQuery
	apiListQuery

	SortField    string `query:"sort_field,oneof=created_at date amount description category_id"`
	SortOrder    string `query:"sort_order,oneof=ASC DESC"`
	CreatedStart string `query:"created_start,integer"`
	CreatedEnd   string `query:"created_end,integer"`
	Q            string `query:"q,max=50"`
	Tag          string `query:"tag,max=50"`
}

// apiExpenseStatsQuery is GET /api/expenses/stats.
type apiExpenseStatsQuery struct {
	apiBoundsQuery

	SortField string `query:"sort_field,oneof=total category"`
	SortOrder string `query:"sort_order,oneof=ASC DESC"`
}

// apiExpenseBudgetsQuery is GET /api/expenses/budgets. Its bounds are
// required: a budget comparison always covers a specific span.
type apiExpenseBudgetsQuery struct {
	Start string `query:"start,integer,required"`
	End   string `query:"end,integer,required"`
	Mode  string `query:"mode,oneof=month months"`
}

// apiDashboardQuery is GET /api/dashboard: the current and prior month's
// bounds.
type apiDashboardQuery struct {
	ThisStart string `query:"this_start,integer,required"`
	ThisEnd   string `query:"this_end,integer,required"`
	LastStart string `query:"last_start,integer,required"`
	LastEnd   string `query:"last_end,integer,required"`
}

// apiRecurrentExpenseListQuery is GET /api/recurrent-expenses.
type apiRecurrentExpenseListQuery struct {
	apiListQuery

	SortField string `query:"sort_field,oneof=created_at description amount period category_id"`
	SortOrder string `query:"sort_order,oneof=ASC DESC"`
	Archived  string `query:"archived,boolean"`
}

// reportQuery is GET /reports/expenses.pdf. It is not token-reachable, so it
// is not in the contract; it goes through decodeQuery only because nothing
// else in the package may read the query string. parseReportMonth checks the
// YYYY-MM format itself, with its own error message.
type reportQuery struct {
	Month string `query:"month"`
}

// queryRule is one parsed `query:"…"` tag.
type queryRule struct {
	key      string
	kind     string // "string", "integer", "positive", "boolean" or "oneof"
	oneOf    []string
	max      int // in characters; 0 means no limit
	required bool
}

// parseQueryRule reads a tag. An unknown rule panics: the tags are constants
// in this package, TestAPIContract parses every one of them, and a typo must
// not quietly turn a rule off.
func parseQueryRule(tag string) queryRule {
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
				panic("query tag " + tag + ": max needs a positive number")
			}

			rule.max = n
		case "required":
			rule.required = true
		default:
			panic("query tag " + tag + ": unknown rule " + name)
		}
	}

	return rule
}

// describe is the rule as contract/api.json records it: "string",
// "string<max 50>", "integer", "positive integer", "boolean" or
// "oneof<a|b>", prefixed "required " when the key must be sent. The mcp module's
// contracttest builds the same strings from its own tags.
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

// check reports the rule a present-or-absent value breaks, or "" when it is
// accepted. The returned name is what the 422's "fields" map carries.
func (rule queryRule) check(value string) string {
	if value == "" {
		if rule.required {
			return "required"
		}

		return ""
	}

	switch rule.kind {
	case "integer":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return "integer"
		}
	case "positive":
		if n, err := strconv.ParseInt(value, 10, 64); err != nil || n < 1 {
			return "positive"
		}
	case "boolean":
		if value != "true" && value != "false" {
			return "boolean"
		}
	case "oneof":
		if !slices.Contains(rule.oneOf, value) {
			return "oneof"
		}
	}

	if rule.max > 0 && utf8.RuneCountInString(value) > rule.max {
		return "max"
	}

	return ""
}

// decodeQuery copies each `query:"…"` key of r's query string into the
// matching string field of dst, a pointer to a struct, and checks it against
// its rule. Every broken rule is collected into one *logic.ValidationError,
// which WriteAPIError answers as a 422 naming each key. Untagged embedded
// structs are walked, as encoding/json does.
func decodeQuery(r *http.Request, dst any) error {
	q := r.URL.Query()
	broken := make(map[string]string)

	walkQuery(reflect.ValueOf(dst).Elem(), func(rule queryRule, field reflect.Value) {
		value := q.Get(rule.key)
		if failed := rule.check(value); failed != "" {
			broken[rule.key] = failed

			return
		}

		field.SetString(value)
	})

	if len(broken) > 0 {
		return logic.NewValidationError(broken)
	}

	return nil
}

// walkQuery calls visit for every tagged string field of v, in declaration
// order, descending into untagged embedded structs.
func walkQuery(v reflect.Value, visit func(queryRule, reflect.Value)) {
	for i := range v.NumField() {
		field := v.Type().Field(i)

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			walkQuery(v.Field(i), visit)

			continue
		}

		tag := field.Tag.Get("query")
		if tag == "" || !field.IsExported() || field.Type.Kind() != reflect.String {
			continue
		}

		visit(parseQueryRule(tag), v.Field(i))
	}
}
