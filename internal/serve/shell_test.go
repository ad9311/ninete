package serve_test

import (
	"html"
	"io"
	"net/http"
	"regexp"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

var (
	metaCSRFTokenRE = regexp.MustCompile(`<meta name="csrf-token" content="([^"]*)"`)
	formCSRFTokenRE = regexp.MustCompile(`name="csrf_token" value="([^"]*)"`)
)

// The SPA's fetch wrapper (web/app/lib/api.ts) reads its CSRF token from a
// <meta> tag in the shell, because nosurf's cookie is HttpOnly and no other
// copy of the token is reachable from JS. Dropping the tag from layout.html
// breaks every non-GET API call at runtime and nothing else, so it is pinned
// here rather than left to a page that happens to render.
func TestLayoutCarriesTheCSRFMetaTag(t *testing.T) {
	s := spec.New(t)

	res := doGet(t, s, "/login", nil)
	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	metaMatches := metaCSRFTokenRE.FindStringSubmatch(string(body))
	require.NotEmpty(t, metaMatches, `<meta name="csrf-token"> is missing from the shell`)

	metaToken := html.UnescapeString(metaMatches[1])
	require.NotEmpty(t, metaToken, "the CSRF meta tag rendered an empty token")

	// The tag must carry the same token the forms post, not a second value that
	// happens to be non-empty.
	formMatches := formCSRFTokenRE.FindStringSubmatch(string(body))
	require.NotEmpty(t, formMatches, "csrf_token not found in the login form")
	require.Equal(t, html.UnescapeString(formMatches[1]), metaToken)
}
