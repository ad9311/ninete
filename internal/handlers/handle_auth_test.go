package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestPostLogout(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "auth_logout_1", "auth_logout_1@example.com", "auth_password_1")
				cookies := s.AuthCookies(t, "auth_logout_1@example.com", "auth_password_1")
				csrfToken, cookies := s.CSRFFrom(t, "/", cookies)

				req := spec.NewPostRequest("/logout", "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, handlers.AppLoginPath, rec.Header().Get("Location"))
			},
		},
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				// Unauthenticated POST to /logout — auth middleware redirects to /login
				csrfToken, cookies := s.CSRFFrom(t, "/login", nil)

				req := spec.NewPostRequest("/logout", "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, handlers.AppLoginPath, rec.Header().Get("Location"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
