package serve

import (
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
)

// The rate limiter is a no-op under ENV=test, so the suite that runs against
// spec.New() cannot tell whether the credential routes actually carry it. This
// test builds a server with a non-test environment instead, which is the only
// way to catch a `root.With(credentialLimit)` going missing from setUpRoutes.
//
// It cannot live in internal/spec: that package imports serve, so a test using
// it from inside serve would close an import cycle.
func newLimitedServer(t *testing.T) *Server {
	t.Helper()

	// A non-test app pulls in chi's access logger, which would print a line per
	// request into the suite's output. Swap it out for the duration.
	previousLogger := middleware.DefaultLogger
	middleware.DefaultLogger = func(next http.Handler) http.Handler { return next }
	t.Cleanup(func() { middleware.DefaultLogger = previousLogger })

	app := &prog.App{
		ENV: prog.ENVDevelopment,
		Logger: prog.NewLogger(prog.LogOptions{
			EnableColor: false,
			EnableQuery: false,
		}),
	}
	require.False(t, app.IsTest(), "the limiter is disabled for a test app")

	sqlDB, err := db.Open()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	server := New(app, logic.New(app, repo.New(app, sqlDB)), sqlDB)
	require.NoError(t, server.LoadTemplates())

	return server
}

var internalCSRFTokenRE = regexp.MustCompile(`name="csrf_token"\s+value="([^"]+)"`)

// csrfFor fetches a form and returns its CSRF token with the cookies that go
// with it. The limiter sits behind the CSRF middleware, so a request without a
// token never reaches it.
func csrfFor(t *testing.T, server *Server, path string) (string, []*http.Cookie) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	matches := internalCSRFTokenRE.FindStringSubmatch(rec.Body.String())
	require.NotEmpty(t, matches, "csrf_token not found in the response for %s", path)

	return html.UnescapeString(matches[1]), rec.Result().Cookies()
}

// postCredentials sends a credential attempt that fails validation before any
// bcrypt work, so exhausting the budget stays cheap.
func postCredentials(
	t *testing.T,
	server *Server,
	path, token string,
	cookies []*http.Cookie,
	remoteAddr string,
) int {
	t.Helper()

	body := url.Values{"email": {""}, "password": {""}}.Encode()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("X-CSRF-Token", token)
	req.RemoteAddr = remoteAddr

	for _, c := range cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	return rec.Code
}

func TestCredentialRoutesAreRateLimited(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			// Deleting either `root.With(credentialLimit)` in setUpRoutes leaves
			// the rest of the suite green, since the limiter is disabled under
			// ENV=test. This is what notices.
			name: "should_throttle_each_credential_route",
			fn: func(t *testing.T) {
				for _, path := range []string{"/login", "/register"} {
					server := newLimitedServer(t)
					token, cookies := csrfFor(t, server, path)

					for i := range authAttemptLimit {
						require.NotEqual(
							t,
							http.StatusTooManyRequests,
							postCredentials(t, server, path, token, cookies, "203.0.113.50:1234"),
							"POST %s attempt %d was throttled before the limit", path, i+1,
						)
					}

					require.Equal(
						t,
						http.StatusTooManyRequests,
						postCredentials(t, server, path, token, cookies, "203.0.113.50:1234"),
						"POST %s was not rate limited", path,
					)
				}
			},
		},
		{
			// The two routes share one middleware value on purpose. Calling
			// authRateLimit() once per route would build two counters and hand
			// a client twice the allowance.
			name: "should_share_one_budget_across_login_and_register",
			fn: func(t *testing.T) {
				server := newLimitedServer(t)
				loginToken, loginCookies := csrfFor(t, server, "/login")
				registerToken, registerCookies := csrfFor(t, server, "/register")

				for range authAttemptLimit {
					require.NotEqual(
						t,
						http.StatusTooManyRequests,
						postCredentials(
							t, server, "/login", loginToken, loginCookies, "203.0.113.51:1234",
						),
					)
				}

				require.Equal(
					t,
					http.StatusTooManyRequests,
					postCredentials(
						t, server, "/register", registerToken, registerCookies, "203.0.113.51:1234",
					),
					"/register had its own budget after /login exhausted one",
				)
			},
		},
		{
			// Rendering the forms must stay free, so only the POSTs are guarded.
			name: "should_not_throttle_rendering_the_forms",
			fn: func(t *testing.T) {
				server := newLimitedServer(t)

				for range authAttemptLimit * 2 {
					req := httptest.NewRequest(http.MethodGet, "/login", nil)
					req.RemoteAddr = "203.0.113.52:1234"
					rec := httptest.NewRecorder()
					server.Router.ServeHTTP(rec, req)

					require.Equal(t, http.StatusOK, rec.Code)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
