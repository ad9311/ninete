package serve

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// TestAPIContractCoversTokenRoutes keeps contract/api.json complete. Every
// route a bearer token can use — on tokenAPIPrefixes, and not DELETE, which no
// token may send — must have an entry, because the contract is what stops the
// MCP server's copies of the JSON shapes from drifting (docs/mcp.md). The list
// behind the file lives in internal/handlers' apiContract; a route added to
// the router without an entry there fails here.
//
// It also fails on an entry for a route the router no longer serves, so the
// file cannot keep describing an endpoint that is gone.
func TestAPIContractCoversTokenRoutes(t *testing.T) {
	data, err := os.ReadFile("contract/api.json")
	require.NoError(t, err)

	var contract map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &contract))

	server := newLimitedServer(t)

	var routes []string

	err = chi.Walk(server.Router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		route = strings.TrimSuffix(route, "/")
		if method == http.MethodDelete || !tokenMayReach(route) {
			return nil
		}

		routes = append(routes, method+" "+route)

		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, routes)

	sort.Strings(routes)

	listed := make([]string, 0, len(contract))
	for route := range contract {
		listed = append(listed, route)
	}

	sort.Strings(listed)

	require.Equal(t, routes, listed,
		"contract/api.json must list exactly the token-reachable routes: add the route to apiContract "+
			"in internal/handlers/api_contract_internal_test.go and run `make contract`")
}
