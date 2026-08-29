package spec

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// WrappedHandler returns the handler Start() serves. Session LoadAndSave now
// lives in the app middleware chain, so the router is already complete.
func (s *Spec) WrappedHandler() http.Handler {
	return s.Server.Router
}

// AuthCookies logs in with real credentials, through the same /api/login the
// SPA itself uses (Phase 7 of docs/spa-migration.md retired the legacy
// form-post login), and returns the cookies needed to make authenticated
// requests.
func (s *Spec) AuthCookies(t *testing.T, email, password string) []*http.Cookie {
	t.Helper()

	csrfToken, cookies := s.CSRFFrom(t, "/login", nil)

	req := NewJSONRequest(http.MethodPost, "/api/login", map[string]string{
		"email":    email,
		"password": password,
	}, cookies, csrfToken)
	rec := httptest.NewRecorder()
	s.WrappedHandler().ServeHTTP(rec, req)

	res := rec.Result()
	defer func() { _ = res.Body.Close() }()

	require.Equal(t, http.StatusNoContent, rec.Code,
		"AuthCookies login failed: %s", rec.Body.String())

	return mergeCookies(cookies, res.Cookies())
}

// csrfTokenRE matches the shell's <meta name="csrf-token"> tag — the only
// place a CSRF token appears in rendered HTML since Phase 7 of
// docs/spa-migration.md deleted every other view, hidden form fields
// included.
var csrfTokenRE = regexp.MustCompile(`name="csrf-token"\s+content="([^"]+)"`)

// CSRFFrom GETs the given URL with cookies, extracts the CSRF token from the
// HTML body, and returns it along with the (possibly updated) cookies.
func (s *Spec) CSRFFrom(t *testing.T, url string, cookies []*http.Cookie) (string, []*http.Cookie) {
	t.Helper()

	req := NewGetRequest(url, cookies)
	rec := httptest.NewRecorder()
	s.WrappedHandler().ServeHTTP(rec, req)

	res := rec.Result()
	defer func() { _ = res.Body.Close() }()

	body := rec.Body.String()
	matches := csrfTokenRE.FindStringSubmatch(body)
	require.NotEmpty(t, matches, "csrf-token not found in response body for %s", url)

	return html.UnescapeString(matches[1]), mergeCookies(cookies, res.Cookies())
}

// NewGetRequest builds a GET request with the given cookies.
func NewGetRequest(url string, cookies []*http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	return req
}

// NewPostRequest builds a POST request with form-encoded body, CSRF header,
// same-origin fetch metadata, and the given cookies.
func NewPostRequest(url, formBody string, cookies []*http.Cookie, csrfToken string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(formBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("X-CSRF-Token", csrfToken)

	for _, c := range cookies {
		req.AddCookie(c)
	}

	return req
}

// NewJSONRequest builds a request with a JSON body, CSRF header, same-origin
// fetch metadata, and the given cookies — the /api/* equivalent of
// NewPostRequest. A nil body sends no body at all, for GET/DELETE calls that
// still need the CSRF and fetch-metadata headers apiCSRF checks.
func NewJSONRequest(method, url string, body any, cookies []*http.Cookie, csrfToken string) *http.Request {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewReader(data)
	}

	req := httptest.NewRequest(method, url, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("X-CSRF-Token", csrfToken)

	for _, c := range cookies {
		req.AddCookie(c)
	}

	return req
}

// mergeCookies merges new cookies into existing ones, replacing by name.
func mergeCookies(existing, newer []*http.Cookie) []*http.Cookie {
	idx := make(map[string]int, len(existing))
	out := make([]*http.Cookie, len(existing))
	copy(out, existing)

	for i, c := range out {
		idx[c.Name] = i
	}

	for _, c := range newer {
		if i, ok := idx[c.Name]; ok {
			out[i] = c
		} else {
			out = append(out, c)
			idx[c.Name] = len(out) - 1
		}
	}

	return out
}
