package serve

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/stretchr/testify/require"
)

// The failure limit is disabled under ENV=test, like authRateLimit, so the
// limiter is exercised directly here.
func TestTokenFailureLimit(t *testing.T) {
	record := func(l *tokenFailureLimit, key string) {
		req := httptest.NewRequest(http.MethodGet, "/api/expenses", nil)
		l.record(httptest.NewRecorder(), req, key)
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_block_a_client_once_it_reaches_the_limit",
			fn: func(t *testing.T) {
				l := newTokenFailureLimit()

				for i := range tokenFailureLimitCount {
					require.False(t, l.blocked("203.0.113.20"), "blocked after %d failures", i)
					record(l, "203.0.113.20")
				}

				require.True(t, l.blocked("203.0.113.20"))
			},
		},
		{
			name: "should_count_each_client_separately",
			fn: func(t *testing.T) {
				l := newTokenFailureLimit()

				for range tokenFailureLimitCount {
					record(l, "203.0.113.21")
				}

				require.False(t, l.blocked("203.0.113.22"))
			},
		},
		{
			name: "should_be_a_no_op_when_disabled",
			fn: func(t *testing.T) {
				var l *tokenFailureLimit

				for range tokenFailureLimitCount + 1 {
					record(l, "203.0.113.23")
				}

				require.False(t, l.blocked("203.0.113.23"))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

func TestTokenMayReach(t *testing.T) {
	allowed := []string{
		"/api/session",
		"/api/expenses",
		"/api/expenses/12",
		"/api/expenses/stats",
		"/api/recurrent-expenses/3/unarchive",
		"/api/report-settings",
	}
	refused := []string{
		"/api/tokens",
		"/api/tokens/1",
		"/api/delete-data",
		"/api/delete-data/expenses",
		"/api/login",
		"/api/register",
		// Segment-matched, so a lookalike prefix does not slip through.
		"/api/expenses-archive",
	}

	for _, path := range allowed {
		require.True(t, tokenMayReach(path), path)
	}

	for _, path := range refused {
		require.False(t, tokenMayReach(path), path)
	}
}

func TestTokenMethodAllowed(t *testing.T) {
	read, write := logic.APITokenScopeRead, logic.APITokenScopeWrite

	require.True(t, tokenMethodAllowed(read, http.MethodGet))
	require.False(t, tokenMethodAllowed(read, http.MethodPost))
	require.False(t, tokenMethodAllowed(read, http.MethodPut))

	require.True(t, tokenMethodAllowed(write, http.MethodGet))
	require.True(t, tokenMethodAllowed(write, http.MethodPost))
	require.True(t, tokenMethodAllowed(write, http.MethodPut))

	for _, scope := range []string{read, write, "admin"} {
		require.False(t, tokenMethodAllowed(scope, http.MethodDelete), scope)
		require.False(t, tokenMethodAllowed(scope, http.MethodPatch), scope)
	}
}
