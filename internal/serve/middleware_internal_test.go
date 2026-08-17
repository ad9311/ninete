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

func TestRealClientIP(t *testing.T) {
	remoteAddrFor := func(t *testing.T, setHeaders func(http.Header)) string {
		t.Helper()

		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "127.0.0.1:41000"
		setHeaders(req.Header)

		var seen string
		realClientIP(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			seen = r.RemoteAddr
		})).ServeHTTP(httptest.NewRecorder(), req)

		return seen
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_use_the_last_forwarded_entry",
			fn: func(t *testing.T) {
				// Caddy appends the address it saw, so the entry it wrote is
				// last. Anything before it came from the client.
				got := remoteAddrFor(t, func(h http.Header) {
					h.Set("X-Forwarded-For", "198.51.100.7, 203.0.113.30")
				})
				require.Equal(t, "203.0.113.30", got)
			},
		},
		{
			name: "should_use_the_last_entry_of_repeated_headers",
			fn: func(t *testing.T) {
				got := remoteAddrFor(t, func(h http.Header) {
					h.Add("X-Forwarded-For", "198.51.100.7")
					h.Add("X-Forwarded-For", "203.0.113.31")
				})
				require.Equal(t, "203.0.113.31", got)
			},
		},
		{
			// Genuine reproduction: chi's middleware.RealIP reads
			// True-Client-IP first, then X-Real-IP, then the *first*
			// X-Forwarded-For entry. Caddy sets neither of the first two and
			// appends to the third, so a client could set all three and mint a
			// fresh rate-limit bucket per request. With RealIP in place this
			// case returned the forged address.
			name: "should_ignore_client_supplied_identity_headers",
			fn: func(t *testing.T) {
				got := remoteAddrFor(t, func(h http.Header) {
					h.Set("True-Client-IP", "198.51.100.1")
					h.Set("X-Real-IP", "198.51.100.2")
					h.Set("X-Forwarded-For", "198.51.100.3, 203.0.113.32")
				})
				require.Equal(t, "203.0.113.32", got)
			},
		},
		{
			name: "should_keep_remote_addr_without_a_forwarded_header",
			fn: func(t *testing.T) {
				got := remoteAddrFor(t, func(http.Header) {})
				require.Equal(t, "127.0.0.1:41000", got)
			},
		},
		{
			name: "should_keep_remote_addr_when_the_last_entry_is_not_an_address",
			fn: func(t *testing.T) {
				got := remoteAddrFor(t, func(h http.Header) {
					h.Set("X-Forwarded-For", "203.0.113.33, not-an-ip")
				})
				require.Equal(t, "127.0.0.1:41000", got)
			},
		},
		{
			name: "should_handle_an_ipv6_entry",
			fn: func(t *testing.T) {
				got := remoteAddrFor(t, func(h http.Header) {
					h.Set("X-Forwarded-For", "198.51.100.7, 2001:db8::2")
				})
				require.Equal(t, "2001:db8::2", got)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

// A forged identity header must not buy a fresh budget. This is the property
// the loopback bind alone does not provide, since the headers arrive through
// the proxy rather than on a direct connection.
func TestAuthRateLimitIgnoresForgedHeaders(t *testing.T) {
	limited := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := realClientIP(newAuthRateLimit(limited)(ok))

	attempt := func(t *testing.T, forged string) int {
		t.Helper()

		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "127.0.0.1:41000"
		req.Header.Set("True-Client-IP", forged)
		req.Header.Set("X-Real-IP", forged)
		req.Header.Set("X-Forwarded-For", forged+", 203.0.113.40")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec.Code
	}

	for i := range authAttemptLimit {
		require.Equal(t, http.StatusOK, attempt(t, fmt.Sprintf("198.51.100.%d", i+1)))
	}

	require.Equal(
		t,
		http.StatusTooManyRequests,
		attempt(t, "198.51.100.200"),
		"a rotating forged header bought a fresh rate-limit budget",
	)
}

func portFor(i int) string {
	return fmt.Sprintf("203.0.113.14:%d", 50000+i)
}
