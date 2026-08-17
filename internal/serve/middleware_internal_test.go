package serve

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// The auth rate limiter is disabled under ENV=test, since the suite logs in far
// more often than a real client ever would. These tests exercise the middleware
// itself so the limit, the key derivation and the throttled response stay
// covered anyway.
func TestAuthRateLimit(t *testing.T) {
	limited := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	post := func(t *testing.T, handler http.Handler, remoteAddr string) int {
		t.Helper()

		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec.Code
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_allow_attempts_up_to_the_limit",
			fn: func(t *testing.T) {
				handler := newAuthRateLimit(limited)(ok)

				for i := range authAttemptLimit {
					require.Equal(
						t,
						http.StatusOK,
						post(t, handler, "203.0.113.10:1234"),
						"attempt %d was throttled before the limit", i+1,
					)
				}
			},
		},
		{
			name: "should_throttle_the_attempt_past_the_limit",
			fn: func(t *testing.T) {
				handler := newAuthRateLimit(limited)(ok)

				for range authAttemptLimit {
					require.Equal(t, http.StatusOK, post(t, handler, "203.0.113.11:1234"))
				}

				require.Equal(t, http.StatusTooManyRequests, post(t, handler, "203.0.113.11:1234"))
			},
		},
		{
			name: "should_bucket_each_client_address_separately",
			fn: func(t *testing.T) {
				handler := newAuthRateLimit(limited)(ok)

				for range authAttemptLimit {
					require.Equal(t, http.StatusOK, post(t, handler, "203.0.113.12:1234"))
				}
				require.Equal(t, http.StatusTooManyRequests, post(t, handler, "203.0.113.12:1234"))

				// A second client must not inherit the first one's exhausted
				// budget. This is what breaks if the key ever collapses to the
				// reverse proxy's own address.
				require.Equal(t, http.StatusOK, post(t, handler, "203.0.113.13:1234"))
			},
		},
		{
			name: "should_ignore_the_source_port_when_keying",
			fn: func(t *testing.T) {
				handler := newAuthRateLimit(limited)(ok)

				// Every attempt arrives on a fresh connection with a new source
				// port. Keying on host:port would hand each one its own budget
				// and never throttle anything.
				for i := range authAttemptLimit {
					require.Equal(t, http.StatusOK, post(t, handler, portFor(i)))
				}

				require.Equal(t, http.StatusTooManyRequests, post(t, handler, portFor(authAttemptLimit)))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestKeyByClientIP(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		expected   string
	}{
		{
			name:       "should_strip_the_port",
			remoteAddr: "203.0.113.20:54321",
			expected:   "203.0.113.20",
		},
		{
			name:       "should_accept_an_address_without_a_port",
			remoteAddr: "203.0.113.21",
			expected:   "203.0.113.21",
		},
		{
			name:       "should_handle_an_ipv6_address",
			remoteAddr: "[2001:db8::1]:54321",
			expected:   "2001:db8::",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login", nil)
			req.RemoteAddr = tc.remoteAddr

			key, err := keyByClientIP(req)
			require.NoError(t, err)
			require.Equal(t, tc.expected, key)
		})
	}
}

func portFor(i int) string {
	return fmt.Sprintf("203.0.113.14:%d", 50000+i)
}
