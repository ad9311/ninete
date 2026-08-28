package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetApp(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, handlers.AppLoginPath, rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_the_shell_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "app_user_1", "app_user_1@example.com", "app_password_1")
				cookies := s.AuthCookies(t, "app_user_1@example.com", "app_password_1")

				req := spec.NewGetRequest("/", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), `<div id="app">`)
				require.Contains(t, rec.Body.String(), `name="csrf-token"`)
				// The manifest-resolved, content-hashed bundle path — never the
				// literal /static/js/build/app.js a stale reference would leave
				// behind.
				require.Contains(t, rec.Body.String(), `src="/static/js/build/app-`)
			},
		},
		{
			// routes/login/Index.svelte and routes/register/Index.svelte need
			// the shell reachable while signed out (Phase 6 of
			// docs/spa-migration.md); AuthMiddleware's guestRoutes carries the
			// exemption.
			name: "should_render_the_shell_at_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/login", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), `<div id="app">`)
			},
		},
		{
			// guestRoutes matches exactly, but root.Get("/*") matches the
			// trailing-slash form too, so a bookmark carrying the slash used to
			// miss the exemption and bounce a guest to the login page they had
			// already asked for.
			name: "should_render_the_shell_at_login_with_a_trailing_slash",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/login/", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), `<div id="app">`)
			},
		},
		{
			name: "should_redirect_away_from_login_when_already_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "app_user_3", "app_user_3@example.com", "app_password_3")
				cookies := s.AuthCookies(t, "app_user_3@example.com", "app_password_3")

				req := spec.NewGetRequest("/login", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, handlers.AppDashboardPath, rec.Header().Get("Location"))
			},
		},
		{
			// The client router owns everything past "/", so a hard refresh on
			// a nested route (docs/spa-migration.md, Phase 1 exit criteria) must
			// resolve to the same shell rather than 404ing.
			name: "should_resolve_a_nested_client_route_on_hard_refresh",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "app_user_2", "app_user_2@example.com", "app_password_2")
				cookies := s.AuthCookies(t, "app_user_2@example.com", "app_password_2")

				req := spec.NewGetRequest("/recurrent-expenses/1/edit", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), `<div id="app">`)
			},
		},
		{
			// The catch-all serves the shell for literally any path — the
			// client router, not the server, decides what "not found" looks
			// like (Phase 7 of docs/spa-migration.md).
			name: "should_render_the_shell_for_an_unknown_path_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "app_user_4", "app_user_4@example.com", "app_password_4")
				cookies := s.AuthCookies(t, "app_user_4@example.com", "app_password_4")

				req := spec.NewGetRequest("/does-not-exist", cookies)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), `<div id="app">`)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
