package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetDashboard(t *testing.T) {
	s := spec.New(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/dashboard", nil)
				rec := httptest.NewRecorder()
				s.WrappedHandler().ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_render_dashboard_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "dash_user_1", "dash_user_1@example.com", "dash_password_1")
				cookies := s.AuthCookies(t, "dash_user_1@example.com", "dash_password_1")

				req := spec.NewGetRequest("/dashboard", cookies)
				rec := httptest.NewRecorder()
				s.WrappedHandler().ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
