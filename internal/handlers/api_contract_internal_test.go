package handlers

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// updateContract regenerates contract/api.json instead of checking it:
//
//	make contract
var updateContract = flag.Bool("update", false, "rewrite contract/api.json from the handler structs") //nolint:gochecknoglobals,lll // test flag

// contractPath is relative to the repository root, where TestMain runs.
const contractPath = "contract/api.json"

// contractEndpoint is one route's entry in contract/api.json: the JSON fields
// of its request body and of its response, flattened to "path": "type".
type contractEndpoint struct {
	Request  map[string]string `json:"request,omitempty"`
	Response map[string]string `json:"response,omitempty"`
}

// apiContract names the struct behind every /api route a bearer token can
// reach (tokenAPIPrefixes in internal/serve) — the routes the MCP server in
// mcp/ may call. A nil side means the route has no JSON body there (a GET's
// request, a 204's response).
//
// TestAPIContractCoversTokenRoutes in internal/serve fails when a
// token-reachable route is missing from the file, so this list cannot quietly
// fall behind the router.
func apiContract() map[string]struct{ request, response any } {
	return map[string]struct{ request, response any }{
		"GET /api/session":    {nil, apiSession{}},
		"GET /api/categories": {nil, apiCategoryListResponse{}},
		"GET /api/dashboard":  {nil, apiDashboardResponse{}},

		"GET /api/report-settings": {nil, apiReportSettingsResponse{}},
		"PUT /api/report-settings": {reportSettingsRequestBody{}, nil},

		"GET /api/tags":        {nil, apiTagListResponse{}},
		"POST /api/tags/retag": {retagRequestBody{}, apiRetagResponse{}},

		"GET /api/expenses":           {nil, apiExpenseListResponse{}},
		"POST /api/expenses":          {expenseRequestBody{}, apiExpenseDetail{}},
		"POST /api/expenses/quick":    {quickExpenseRequestBody{}, apiExpenseDetail{}},
		"GET /api/expenses/stats":     {nil, apiExpenseStatsResponse{}},
		"GET /api/expenses/budgets":   {nil, apiExpenseBudgetsResponse{}},
		"PUT /api/expenses/budgets":   {expenseBudgetsRequestBody{}, nil},
		"GET /api/expenses/{id}":      {nil, apiExpenseDetail{}},
		"PUT /api/expenses/{id}":      {expenseRequestBody{}, apiExpenseDetail{}},
		"GET /api/recurrent-expenses": {nil, apiRecurrentExpenseListResponse{}},

		"POST /api/recurrent-expenses":                {recurrentExpenseRequestBody{}, apiRecurrentExpense{}},
		"GET /api/recurrent-expenses/{id}":            {nil, apiRecurrentExpense{}},
		"PUT /api/recurrent-expenses/{id}":            {recurrentExpenseRequestBody{}, apiRecurrentExpense{}},
		"POST /api/recurrent-expenses/{id}/unarchive": {nil, apiRecurrentExpense{}},
	}
}

// TestAPIContract keeps contract/api.json equal to what the handlers actually
// encode and decode. The MCP server's own test (mcp/internal/api) checks its
// copies of these shapes against the same file, so a JSON change here fails
// that side until mcp/ is updated too. See docs/mcp.md, "Keeping the API
// contract".
func TestAPIContract(t *testing.T) {
	want := make(map[string]contractEndpoint)

	for route, types := range apiContract() {
		var endpoint contractEndpoint

		if types.request != nil {
			endpoint.Request = contractFields(reflect.TypeOf(types.request))
		}

		if types.response != nil {
			endpoint.Response = contractFields(reflect.TypeOf(types.response))
		}

		want[route] = endpoint
	}

	// SetEscapeHTML(false) keeps "array<object>" readable in the diff rather
	// than "array\u003cobject\u003e". Encode sorts map keys, so the file is
	// deterministic.
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	require.NoError(t, enc.Encode(want))

	encoded := buf.Bytes()

	if *updateContract {
		require.NoError(t, os.MkdirAll("contract", 0o750))
		require.NoError(t, os.WriteFile(contractPath, encoded, 0o600))

		return
	}

	current, err := os.ReadFile(contractPath)
	require.NoError(t, err, "run `make contract` to create %s", contractPath)
	require.Equal(t, string(encoded), string(current),
		"%s is out of date with the handler structs: run `make contract`, then update mcp/ to match",
		contractPath)
}

// contractFields flattens a type's JSON encoding to "path": "type" pairs, the
// way encoding/json sees it: json tags name the fields, "-" drops one, and an
// untagged embedded struct is inlined. A slice element's fields sit under
// "name[]", a nested object's under "name.".
//
// mcp/internal/contracttest carries the same walk for the MCP side's types.
// The two are kept in step by the file itself: if they disagreed, one of the
// two tests would fail on its first run.
func contractFields(t reflect.Type) map[string]string {
	fields := make(map[string]string)
	walkContract(t, "", fields)

	return fields
}

func walkContract(t reflect.Type, prefix string, fields map[string]string) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	for field := range t.Fields() {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}

		if field.Anonymous && name == "" && field.Type.Kind() == reflect.Struct {
			walkContract(field.Type, prefix, fields)

			continue
		}

		if !field.IsExported() {
			continue
		}

		if name == "" {
			name = field.Name
		}

		describeContract(field.Type, prefix+name, fields)
	}
}

func describeContract(t reflect.Type, path string, fields map[string]string) {
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
		walkContract(t, path+".", fields)
	case reflect.Slice, reflect.Array:
		elem := t.Elem()
		if elem.Kind() == reflect.Struct {
			fields[path] = "array<object>" + suffix
			walkContract(elem, path+"[].", fields)

			return
		}

		fields[path] = "array<" + contractScalar(elem) + ">" + suffix
	case reflect.Map:
		fields[path] = "map<" + contractScalar(t.Elem()) + ">" + suffix
	default:
		fields[path] = contractScalar(t) + suffix
	}
}

func contractScalar(t reflect.Type) string {
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
