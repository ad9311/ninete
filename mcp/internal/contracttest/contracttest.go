// Package contracttest checks this module's copies of the API's JSON shapes
// against contract/api.json, the file the app's own test suite generates from
// its handler structs (`make contract`). It is imported only by tests.
//
// The rules, and why they differ by direction:
//   - Request bodies must match the server's exactly. The server's decoder
//     ignores unknown fields, and a PUT replaces the whole record, so a field
//     this module does not send is silently reset — the data-loss case.
//   - Responses may be a subset. This module may ignore a field the server
//     sends, but every field it reads must exist with the same JSON type.
//   - Query keys may be a subset, but each one this module declares must exist
//     on the server with the same rule — a oneof may list fewer values, never
//     one the server lacks — and every key the server requires must be
//     declared here as required (QueryProblems). The server answers any value
//     it does not accept with a 422, so this catches before release what would
//     otherwise surface as a failing tool call.
package contracttest

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/ad9311/ninete-mcp/internal/api"
)

// Endpoint names this module's types for a route: the query struct the tool
// encodes (api/query.go), the request body and the response. A nil side means
// this module sends no query or body there, or reads none.
type Endpoint struct {
	Query    any
	Request  any
	Response any
}

// Endpoints lists every route the tools call, with the types they use for it.
// The tools test fails when a tool calls a route missing from this list, so a
// new call cannot skip the check.
func Endpoints() map[string]Endpoint {
	return map[string]Endpoint{
		"GET /api/categories": {nil, nil, api.CategoryList{}},
		"GET /api/dashboard":  {api.DashboardQuery{}, nil, api.Dashboard{}},

		"GET /api/report-settings": {nil, nil, api.ReportSettings{}},
		"PUT /api/report-settings": {nil, api.ReportSettingsBody{}, nil},

		"GET /api/tags":        {nil, nil, api.TagList{}},
		"POST /api/tags/retag": {nil, api.RetagBody{}, api.Retag{}},

		"GET /api/expenses":         {api.ExpenseListQuery{}, nil, api.ExpenseList{}},
		"POST /api/expenses":        {nil, api.ExpenseBody{}, api.Expense{}},
		"POST /api/expenses/quick":  {nil, api.QuickExpenseBody{}, api.Expense{}},
		"GET /api/expenses/stats":   {api.ExpenseStatsQuery{}, nil, api.ExpenseStats{}},
		"GET /api/expenses/budgets": {api.BudgetsQuery{}, nil, api.Budgets{}},
		"PUT /api/expenses/budgets": {nil, api.BudgetsBody{}, nil},
		"GET /api/expenses/{id}":    {nil, nil, api.Expense{}},
		"PUT /api/expenses/{id}":    {nil, api.ExpenseBody{}, api.Expense{}},

		"GET /api/recurrent-expenses":                 {api.RecurrentExpenseListQuery{}, nil, api.RecurrentExpenseList{}},
		"POST /api/recurrent-expenses":                {nil, api.RecurrentExpenseBody{}, api.RecurrentExpense{}},
		"GET /api/recurrent-expenses/{id}":            {nil, nil, api.RecurrentExpense{}},
		"PUT /api/recurrent-expenses/{id}":            {nil, api.RecurrentExpenseBody{}, api.RecurrentExpense{}},
		"POST /api/recurrent-expenses/{id}/unarchive": {nil, nil, api.RecurrentExpense{}},
	}
}

// Contract is one route's entry in contract/api.json.
type Contract struct {
	Query    map[string]string `json:"query"`
	Request  map[string]string `json:"request"`
	Response map[string]string `json:"response"`
}

// Load reads contract/api.json from the repository root, found relative to
// this source file so the tests work from any directory.
func Load() (map[string]Contract, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return nil, ErrNoSource
	}

	// mcp/internal/contracttest/contracttest.go -> repository root.
	path := filepath.Join(filepath.Dir(file), "..", "..", "..", "contract", "api.json")

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("contracttest: %w", err)
	}

	var contract map[string]Contract
	if err := json.Unmarshal(data, &contract); err != nil {
		return nil, fmt.Errorf("contracttest: %w", err)
	}

	return contract, nil
}

var idSegment = regexp.MustCompile(`/\d+(/|$)`)

// Route turns a concrete request ("GET", "/api/expenses/7") into its contract
// key ("GET /api/expenses/{id}").
func Route(method, path string) string {
	return method + " " + idSegment.ReplaceAllString(path, "/{id}$1")
}

// UnknownQueryKeys returns the keys of query that the server does not read on
// route, sorted. The tools test runs it over every request a tool sends — a
// runtime backstop to QueryProblems, which checks the declarations.
func UnknownQueryKeys(contract Contract, query url.Values) []string {
	var unknown []string

	for key := range query {
		if _, ok := contract.Query[key]; !ok {
			unknown = append(unknown, key)
		}
	}

	slices.Sort(unknown)

	return unknown
}

// QueryProblems compares this module's query rules for a route (api.QueryRules)
// with the server's, and describes every disagreement, sorted:
//   - a key the server does not read;
//   - a rule that differs, except a oneof listing a subset of the server's;
//   - a key the server requires that is not declared required here.
func QueryProblems(server, mine map[string]string) []string {
	var problems []string

	for key, rule := range mine {
		serverRule, ok := server[key]

		switch {
		case !ok:
			problems = append(problems, fmt.Sprintf("%s: the server does not read it", key))
		case !ruleFits(rule, serverRule):
			problems = append(problems, fmt.Sprintf("%s: declared %q, the server expects %q", key, rule, serverRule))
		}
	}

	for key, serverRule := range server {
		if !strings.HasPrefix(serverRule, "required ") {
			continue
		}

		if rule, ok := mine[key]; !ok || !strings.HasPrefix(rule, "required ") {
			problems = append(problems, fmt.Sprintf("%s: the server requires it", key))
		}
	}

	slices.Sort(problems)

	return problems
}

// ruleFits reports whether a value meeting mine always meets server. Being
// required here when the server is not is fine: it only means always sending.
func ruleFits(mine, server string) bool {
	mine = strings.TrimPrefix(mine, "required ")
	server = strings.TrimPrefix(server, "required ")

	mineValues, mineIsOneOf := oneOfValues(mine)
	serverValues, serverIsOneOf := oneOfValues(server)

	if mineIsOneOf && serverIsOneOf {
		for _, value := range mineValues {
			if !slices.Contains(serverValues, value) {
				return false
			}
		}

		return true
	}

	return mine == server
}

func oneOfValues(rule string) ([]string, bool) {
	inner, ok := strings.CutPrefix(rule, "oneof<")
	if !ok {
		return nil, false
	}

	return strings.Split(strings.TrimSuffix(inner, ">"), "|"), true
}

// Fields flattens a value's JSON encoding to "path": "type" pairs — the same
// walk the app's TestAPIContract does over the handler structs, so the two
// sides are comparable. See internal/handlers/api_contract_internal_test.go.
func Fields(v any) map[string]string {
	fields := make(map[string]string)
	walk(reflect.TypeOf(v), "", fields)

	return fields
}

func walk(t reflect.Type, prefix string, fields map[string]string) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	for field := range t.Fields() {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}

		if field.Anonymous && name == "" && field.Type.Kind() == reflect.Struct {
			walk(field.Type, prefix, fields)

			continue
		}

		if !field.IsExported() {
			continue
		}

		if name == "" {
			name = field.Name
		}

		describe(field.Type, prefix+name, fields)
	}
}

func describe(t reflect.Type, path string, fields map[string]string) {
	nullable := false
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
		nullable = true
	}

	suffix := ""
	if nullable {
		suffix = "|null"
	}

	switch t.Kind() {
	case reflect.Struct:
		fields[path] = "object" + suffix
		walk(t, path+".", fields)
	case reflect.Slice, reflect.Array:
		elem := t.Elem()
		if elem.Kind() == reflect.Struct {
			fields[path] = "array<object>" + suffix
			walk(elem, path+"[].", fields)

			return
		}

		fields[path] = "array<" + scalar(elem) + ">" + suffix
	case reflect.Map:
		fields[path] = "map<" + scalar(t.Elem()) + ">" + suffix
	default:
		fields[path] = scalar(t) + suffix
	}
}

func scalar(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	default:
		return t.Kind().String()
	}
}
