package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetRoot(t *testing.T) {
	s := spec.New(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_login_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/", nil)
				rec := httptest.NewRecorder()
				s.WrappedHandler().ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_redirect_to_dashboard_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "root_user_1", "root_user_1@example.com", "root_password_1")
				cookies := s.AuthCookies(t, "root_user_1@example.com", "root_password_1")

				req := spec.NewGetRequest("/", cookies)
				rec := httptest.NewRecorder()
				s.WrappedHandler().ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/dashboard", rec.Header().Get("Location"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
