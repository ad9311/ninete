package serve_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

// doAPI sends a request through the real router and decodes the error envelope.
// Sec-Fetch-Site is set for the same reason internal/spec sets it on form posts:
// nosurf's ensureSameOrigin runs before the token is looked at, and a
// hand-built request carries none of the metadata a browser sends.
func doAPI(
	t *testing.T,
	s spec.Spec,
	method, path string,
	cookies []*http.Cookie,
	csrfToken string,
) (*http.Response, map[string]any) {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	s.WrappedHandler().ServeHTTP(rec, req)

	res := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, res.Body.Close())
	})

	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body),
		"API response was not JSON: %s", rec.Body.String())

	return res, body
}

func TestAPIRoutesUseTheirOwnChain(t *testing.T) {
	s := spec.New(t)

	user := s.CreateAuthUser(t, "api_chain_user", "api_chain@test.com", "Password123!")
	require.NotZero(t, user.ID)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			// The whole point of the separate chain. Under the HTML
			// AuthMiddleware this would be a 303 to /login, which fetch follows
			// silently and hands the caller as HTML under status 200.
			name: "should_answer_a_signed_out_request_with_401_json_and_no_location",
			fn: func(t *testing.T) {
				res, body := doAPI(t, s, http.MethodGet, "/api/session", nil, "")

				require.Equal(t, http.StatusUnauthorized, res.StatusCode)
				require.Empty(t, res.Header.Get("Location"), "the API chain redirected")
				require.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))
				require.Equal(t, "authentication required", body["error"])
			},
		},
		{
			// The API chain drops setTmplData, which is the only other place
			// KeyCurrentUser is set, and getCurrentUser panics when the key is
			// absent. Without apiAuth filling it, this is a 500 that looks like
			// a client bug.
			name: "should_put_the_signed_in_user_into_the_request_context",
			fn: func(t *testing.T) {
				cookies := s.AuthCookies(t, "api_chain@test.com", "Password123!")
				res, body := doAPI(t, s, http.MethodGet, "/api/session", cookies, "")

				require.Equal(t, http.StatusOK, res.StatusCode)
				require.Equal(t, "api_chain_user", body["username"])
				require.Equal(t, "api_chain@test.com", body["email"])
				require.NotContains(t, body, "passwordHash")
			},
		},
		{
			// Reaching the 404 proves apiAuth let the request through, so the
			// FindUser lookup that replaces setTmplData ran cleanly.
			name: "should_answer_a_signed_in_request_to_an_unknown_route_with_404_json",
			fn: func(t *testing.T) {
				cookies := s.AuthCookies(t, "api_chain@test.com", "Password123!")
				res, body := doAPI(t, s, http.MethodGet, "/api/not-a-resource", cookies, "")

				require.Equal(t, http.StatusNotFound, res.StatusCode)
				require.Equal(t, "resource not found", body["error"])
			},
		},
		{
			name: "should_reject_an_unsafe_request_without_a_csrf_token",
			fn: func(t *testing.T) {
				cookies := s.AuthCookies(t, "api_chain@test.com", "Password123!")
				res, body := doAPI(t, s, http.MethodPost, "/api/session", cookies, "")

				require.Equal(t, http.StatusForbidden, res.StatusCode)
				require.Equal(t, "invalid or missing CSRF token", body["error"])
			},
		},
		{
			// The API shares the page chain's CSRF cookie, so a token minted
			// while rendering a page is the token the API accepts.
			name: "should_accept_an_unsafe_request_carrying_the_page_csrf_token",
			fn: func(t *testing.T) {
				cookies := s.AuthCookies(t, "api_chain@test.com", "Password123!")
				token, cookies := s.CSRFFrom(t, "/foods/new", cookies)

				res, _ := doAPI(t, s, http.MethodPost, "/api/session", cookies, token)

				// 405 rather than 403: CSRF passed and the request reached
				// routing, where /api/session has no POST.
				require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode,
					"a valid page token was rejected by the API chain")
			},
		},
		{
			// A JSON response constrains no document, so the API chain skips
			// contentSecurityPolicy. The base headers still apply.
			name: "should_skip_the_csp_but_keep_the_base_security_headers",
			fn: func(t *testing.T) {
				res, _ := doAPI(t, s, http.MethodGet, "/api/session", nil, "")

				require.Empty(t, res.Header.Get("Content-Security-Policy"))
				require.Equal(t, "nosniff", res.Header.Get("X-Content-Type-Options"))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}
