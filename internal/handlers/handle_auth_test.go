package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGetLogin(t *testing.T) {
	s := spec.New(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_login_page_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/login", nil)
				rec := httptest.NewRecorder()
				s.WrappedHandler().ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "csrf_token")
			},
		},
		{
			name: "should_redirect_to_dashboard_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "auth_login_get_1", "auth_login_get_1@example.com", "auth_password_1")
				cookies := s.AuthCookies(t, "auth_login_get_1@example.com", "auth_password_1")

				req := spec.NewGetRequest("/login", cookies)
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

func TestGetRegister(t *testing.T) {
	s := spec.New(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_render_register_page_when_unauthenticated",
			fn: func(t *testing.T) {
				req := spec.NewGetRequest("/register", nil)
				rec := httptest.NewRecorder()
				s.WrappedHandler().ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				require.Contains(t, rec.Body.String(), "csrf_token")
			},
		},
		{
			name: "should_redirect_to_dashboard_when_authenticated",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "auth_register_get_1", "auth_register_get_1@example.com", "auth_password_1")
				cookies := s.AuthCookies(t, "auth_register_get_1@example.com", "auth_password_1")

				req := spec.NewGetRequest("/register", cookies)
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

func TestPostLogin(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_root_with_valid_credentials",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "auth_post_login_1", "auth_post_login_1@example.com", "auth_password_1")

				csrfToken, cookies := s.CSRFFrom(t, "/login", nil)

				form := url.Values{
					"email":    {"auth_post_login_1@example.com"},
					"password": {"auth_password_1"},
				}
				req := spec.NewPostRequest("/login", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_bad_request_with_wrong_password",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "auth_post_login_2", "auth_post_login_2@example.com", "auth_password_2")

				csrfToken, cookies := s.CSRFFrom(t, "/login", nil)

				form := url.Values{
					"email":    {"auth_post_login_2@example.com"},
					"password": {"wrong_password"},
				}
				req := spec.NewPostRequest("/login", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostRegister(t *testing.T) {
	s := spec.New(t)
	handler := s.WrappedHandler()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_redirect_to_dashboard_with_valid_signup",
			fn: func(t *testing.T) {
				s.CreateInvitationCode(t, "handler_invite_1")

				csrfToken, cookies := s.CSRFFrom(t, "/register", nil)

				form := url.Values{
					"username":             {"authreguser1"},
					"email":                {"auth_reg_user_1@example.com"},
					"password":             {"auth_reg_password_1"},
					"passwordConfirmation": {"auth_reg_password_1"},
					"invitationCode":       {"handler_invite_1"},
				}
				req := spec.NewPostRequest("/register", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/dashboard", rec.Header().Get("Location"))
			},
		},
		{
			name: "should_return_bad_request_with_invalid_invite_code",
			fn: func(t *testing.T) {
				csrfToken, cookies := s.CSRFFrom(t, "/register", nil)

				form := url.Values{
					"username":             {"authreguser2"},
					"email":                {"auth_reg_user_2@example.com"},
					"password":             {"auth_reg_password_2"},
					"passwordConfirmation": {"auth_reg_password_2"},
					"invitationCode":       {"invalid_invite_code"},
				}
				req := spec.NewPostRequest("/register", form.Encode(), cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

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
				csrfToken, cookies := s.CSRFFrom(t, "/dashboard", cookies)

				req := spec.NewPostRequest("/logout", "", cookies, csrfToken)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.Equal(t, "/login", rec.Header().Get("Location"))
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
				require.Contains(t, rec.Header().Get("Location"), "/login")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
