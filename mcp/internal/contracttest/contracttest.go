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
package contracttest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/ad9311/ninete-mcp/internal/api"
)

// Endpoint names this module's type for each side of a route. A nil side
// means this module sends no body there, or reads none.
type Endpoint struct {
	Request  any
	Response any
}

// Endpoints lists every route the tools call, with the types they use for it.
// The tools test fails when a tool calls a route missing from this list, so a
// new call cannot skip the check.
func Endpoints() map[string]Endpoint {
	return map[string]Endpoint{
		"GET /api/categories": {nil, api.CategoryList{}},
		"GET /api/dashboard":  {nil, api.Dashboard{}},

		"GET /api/report-settings": {nil, api.ReportSettings{}},
		"PUT /api/report-settings": {api.ReportSettingsBody{}, nil},

		"GET /api/expenses":         {nil, api.ExpenseList{}},
		"POST /api/expenses":        {api.ExpenseBody{}, api.Expense{}},
		"POST /api/expenses/quick":  {api.QuickExpenseBody{}, api.Expense{}},
		"GET /api/expenses/stats":   {nil, api.ExpenseStats{}},
		"GET /api/expenses/budgets": {nil, api.Budgets{}},
		"PUT /api/expenses/budgets": {api.BudgetsBody{}, nil},
		"GET /api/expenses/{id}":    {nil, api.Expense{}},
		"PUT /api/expenses/{id}":    {api.ExpenseBody{}, api.Expense{}},

		"GET /api/recurrent-expenses":                 {nil, api.RecurrentExpenseList{}},
		"POST /api/recurrent-expenses":                {api.RecurrentExpenseBody{}, api.RecurrentExpense{}},
		"GET /api/recurrent-expenses/{id}":            {nil, api.RecurrentExpense{}},
		"PUT /api/recurrent-expenses/{id}":            {api.RecurrentExpenseBody{}, api.RecurrentExpense{}},
		"POST /api/recurrent-expenses/{id}/unarchive": {nil, api.RecurrentExpense{}},
	}
}

// Contract is one route's entry in contract/api.json.
type Contract struct {
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
