package serve_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

// bundlePath reads the manifest the same way internal/serve does, so the test
// checks a real deployed filename rather than a literal that would go stale
// the moment web/build.ts's hashing changes it.
func bundlePath(t *testing.T, entry string) string {
	t.Helper()

	data, err := os.ReadFile("./web/static/js/build/manifest.json")
	require.NoError(t, err)

	var manifest map[string]string
	require.NoError(t, json.Unmarshal(data, &manifest))

	name, ok := manifest[entry]
	require.True(t, ok, "manifest missing entry %q", entry)

	return "/static/js/build/" + name
}

func hasCookie(res *http.Response, name string) bool {
	for _, c := range res.Cookies() {
		if c.Name == name {
			return true
		}
	}

	return false
}

func doGet(t *testing.T, s spec.Spec, url string, cookies []*http.Cookie) *http.Response {
	t.Helper()

	req := spec.NewGetRequest(url, cookies)
	rec := httptest.NewRecorder()
	s.WrappedHandler().ServeHTTP(rec, req)

	res := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, res.Body.Close())
	})

	return res
}

func TestStaticAssetsBypassAppMiddleware(t *testing.T) {
	s := spec.New(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_not_issue_session_or_csrf_cookies_for_assets",
			fn: func(t *testing.T) {
				res := doGet(t, s, "/static/css/layout.css", nil)

				require.Equal(t, http.StatusOK, res.StatusCode)
				require.False(t, hasCookie(res, "ninete_session"), "asset request loaded a session")
				require.False(t, hasCookie(res, "ninete_csrf"), "asset request issued a CSRF token")
				require.Empty(t, res.Header.Get("Content-Security-Policy"), "asset request built a CSP nonce")
			},
		},
		{
			name: "should_send_cache_and_base_security_headers_for_assets",
			fn: func(t *testing.T) {
				res := doGet(t, s, "/static/css/layout.css", nil)

				require.Equal(t, "public, max-age=300", res.Header.Get("Cache-Control"))
				require.Equal(t, "nosniff", res.Header.Get("X-Content-Type-Options"))
			},
		},
		{
			// staticCacheHeaders sets Cache-Control before the file server runs,
			// so a 404 would inherit it and outlive the deploy that ships the
			// asset. It does not, because ServeContent strips Cache-Control on
			// its error paths — a property of net/http, not of our code, which
			// is why it is pinned here.
			name: "should_not_mark_a_missing_asset_as_cacheable",
			fn: func(t *testing.T) {
				res := doGet(t, s, "/static/css/missing.css", nil)

				require.Equal(t, http.StatusNotFound, res.StatusCode)
				require.Empty(t, res.Header.Get("Cache-Control"), "a 404 was marked cacheable")
			},
		},
		{
			name: "should_keep_the_cache_header_on_a_revalidated_asset",
			fn: func(t *testing.T) {
				res := doGet(t, s, "/static/css/layout.css", nil)
				require.Equal(t, http.StatusOK, res.StatusCode)

				lastModified := res.Header.Get("Last-Modified")
				require.NotEmpty(t, lastModified)

				req := spec.NewGetRequest("/static/css/layout.css", nil)
				req.Header.Set("If-Modified-Since", lastModified)
				rec := httptest.NewRecorder()
				s.WrappedHandler().ServeHTTP(rec, req)

				require.Equal(t, http.StatusNotModified, rec.Code)
				require.Equal(t, "public, max-age=300", rec.Header().Get("Cache-Control"))
			},
		},
		{
			name: "should_serve_assets_without_authentication",
			fn: func(t *testing.T) {
				res := doGet(t, s, bundlePath(t, "index"), nil)

				require.Equal(t, http.StatusOK, res.StatusCode)
			},
		},
		{
			name: "should_still_apply_the_app_chain_to_pages",
			fn: func(t *testing.T) {
				res := doGet(t, s, "/login", nil)

				require.Equal(t, http.StatusOK, res.StatusCode)
				require.True(t, hasCookie(res, "ninete_csrf"), "page request skipped CSRF")
				require.NotEmpty(t, res.Header.Get("Content-Security-Policy"), "page request skipped CSP")
				require.Empty(t, res.Header.Get("Cache-Control"), "page request was marked cacheable")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

func TestNotFoundKeepsTemplateData(t *testing.T) {
	s := spec.New(t)
	s.CreateAuthUser(t, "routes_user_1", "routes_user_1@example.com", "routes_password_1")
	cookies := s.AuthCookies(t, "routes_user_1@example.com", "routes_password_1")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			// The fallbacks are registered on the middleware group precisely so
			// they still receive template data; without it render panics.
			name: "should_render_the_not_found_page_for_unknown_paths",
			fn: func(t *testing.T) {
				res := doGet(t, s, "/does-not-exist", cookies)

				require.Equal(t, http.StatusNotFound, res.StatusCode)
				require.Contains(t, res.Header.Get("Content-Type"), "text/html")
			},
		},
		{
			name: "should_return_a_plain_404_for_missing_assets",
			fn: func(t *testing.T) {
				res := doGet(t, s, "/static/css/missing.css", nil)

				require.Equal(t, http.StatusNotFound, res.StatusCode)
				require.False(t, hasCookie(res, "ninete_session"), "missing asset loaded a session")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}
