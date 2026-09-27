package api_test

import (
	"testing"

	"github.com/ad9311/ninete-mcp/internal/contracttest"
	"github.com/stretchr/testify/require"
)

// TestContract checks this module's JSON types against contract/api.json,
// which the app's suite generates from its handler structs. When the app
// changes a route's JSON and regenerates the file, this fails until the types
// here follow. The rules are on the contracttest package.
func TestContract(t *testing.T) {
	contract, err := contracttest.Load()
	require.NoError(t, err)

	for route, endpoint := range contracttest.Endpoints() {
		t.Run(route, func(t *testing.T) {
			server, ok := contract[route]
			require.True(t, ok, "%s is not in contract/api.json: the app no longer serves it to tokens", route)

			if endpoint.Request == nil {
				require.Empty(t, server.Request, "the server now expects a request body")
			} else {
				// Exact: a field the server reads but this module does not send
				// would be reset on every update.
				require.Equal(t, server.Request, contracttest.Fields(endpoint.Request),
					"request body differs from the server's")
			}

			if endpoint.Response == nil {
				return
			}

			for path, kind := range contracttest.Fields(endpoint.Response) {
				serverKind, ok := server.Response[path]
				require.True(t, ok, "response field %q is not sent by the server", path)
				require.Equal(t, serverKind, kind, "response field %q has a different type", path)
			}
		})
	}
}

// TestContractCatchesDrift proves the check is not vacuous: a body missing a
// field the server reads, and a response reading a field the server does not
// send, are both caught by the same comparison TestContract makes.
func TestContractCatchesDrift(t *testing.T) {
	contract, err := contracttest.Load()
	require.NoError(t, err)

	server := contract["PUT /api/expenses/{id}"]

	type bodyWithoutNote struct {
		CategoryID  int      `json:"category_id"`
		Description string   `json:"description"`
		Amount      uint64   `json:"amount"`
		Date        int64    `json:"date"`
		Tags        []string `json:"tags"`
	}

	require.NotEqual(t, server.Request, contracttest.Fields(bodyWithoutNote{}))

	type listReadingNote struct {
		Data []struct {
			Note string `json:"note"`
		} `json:"data"`
	}

	_, sent := contract["GET /api/expenses"].Response["data[].note"]
	require.False(t, sent, "the list must not send a note")
	require.Contains(t, contracttest.Fields(listReadingNote{}), "data[].note")
}
