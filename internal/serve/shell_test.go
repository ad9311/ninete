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

var metaCSRFTokenRE = regexp.MustCompile(`<meta name="csrf-token" content="([^"]*)"`)

// The SPA's fetch wrapper (web/app/lib/api.ts) reads its CSRF token from a
// <meta> tag in the shell, because nosurf's cookie is HttpOnly and no other
// copy of the token is reachable from JS. Dropping the tag from the shell
// breaks every non-GET API call at runtime and nothing else, so it is pinned
// here rather than left to a page that happens to render. Since Phase 7 of
// docs/spa-migration.md this tag is the only place a CSRF token appears in
// rendered HTML — there is no more hidden form field to cross-check it
// against.
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
}
