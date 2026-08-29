package serve

import (
	"encoding/json"
	"html"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/handlers"
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
	require.NoError(t, server.LoadAssetManifest())

	return server
}

var internalCSRFTokenRE = regexp.MustCompile(`name="csrf-token"\s+content="([^"]+)"`)

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

// postAPICredentials sends a credential attempt that fails validation before
// any bcrypt work, so exhausting the budget stays cheap.
func postAPICredentials(
	t *testing.T,
	server *Server,
	path, token string,
	cookies []*http.Cookie,
	remoteAddr string,
) (int, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"email":"","password":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("X-CSRF-Token", token)
	req.RemoteAddr = remoteAddr

	for _, c := range cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	return rec.Code, rec.Body.String()
}

func TestCredentialRoutesAreRateLimited(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			// Deleting either `api.With(credentialLimit)` in setUpAPIRoutes
			// leaves the rest of the suite green, since the limiter is
			// disabled under ENV=test. This is what notices.
			name: "should_throttle_each_credential_route",
			fn: func(t *testing.T) {
				for _, path := range []string{"/api/login", "/api/register"} {
					server := newLimitedServer(t)
					token, cookies := csrfFor(t, server, "/login")

					for i := range authAttemptLimit {
						code, _ := postAPICredentials(t, server, path, token, cookies, "203.0.113.50:1234")
						require.NotEqual(
							t, http.StatusTooManyRequests, code,
							"POST %s attempt %d was throttled before the limit", path, i+1,
						)
					}

					code, _ := postAPICredentials(t, server, path, token, cookies, "203.0.113.50:1234")
					require.Equal(t, http.StatusTooManyRequests, code, "POST %s was not rate limited", path)
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
				token, cookies := csrfFor(t, server, "/login")

				for range authAttemptLimit {
					code, _ := postAPICredentials(t, server, "/api/login", token, cookies, "203.0.113.51:1234")
					require.NotEqual(t, http.StatusTooManyRequests, code)
				}

				code, _ := postAPICredentials(t, server, "/api/register", token, cookies, "203.0.113.51:1234")
				require.Equal(
					t, http.StatusTooManyRequests, code,
					"/api/register had its own budget after /api/login exhausted one",
				)
			},
		},
		{
			// Genuine reproduction: TooManyRequests renders the HTML error
			// page through tmplData, which panics on the API chain because
			// setUpAPIMiddlewares drops setTmplData — Recoverer turned that
			// into an empty 500 instead of a 429 with the envelope.
			name: "should_answer_a_throttled_api_credential_route_with_the_json_envelope",
			fn: func(t *testing.T) {
				server := newLimitedServer(t)
				token, cookies := csrfFor(t, server, "/login")

				for range authAttemptLimit {
					code, _ := postAPICredentials(
						t, server, "/api/login", token, cookies, "203.0.113.54:1234",
					)
					require.NotEqual(t, http.StatusTooManyRequests, code)
				}

				code, body := postAPICredentials(
					t, server, "/api/login", token, cookies, "203.0.113.54:1234",
				)
				require.Equal(t, http.StatusTooManyRequests, code)

				var errBody struct {
					Error string `json:"error"`
				}
				require.NoError(t, json.Unmarshal([]byte(body), &errBody))
				require.Equal(t, handlers.ErrTooManyAttempts.Error(), errBody.Error)
			},
		},
		{
			// Rendering the shell must stay free, so only the API POSTs are
			// guarded.
			name: "should_not_throttle_rendering_the_shell",
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
