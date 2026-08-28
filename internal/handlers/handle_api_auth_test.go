package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

// guestJSON is doJSON for a request with no session yet: CSRFFrom mints a
// token off a guest-reachable page, since nosurf's cookie is shared by both
// chains regardless of auth state.
func guestJSON(
	t *testing.T,
	s spec.Spec,
	method, url string,
	body any,
) (*http.Response, []byte) {
	t.Helper()

	csrfToken, cookies := s.CSRFFrom(t, "/login", nil)

	return doJSON(t, s.WrappedHandler(), method, url, body, cookies, csrfToken)
}

type apiSessionBody struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func TestAPILogin(t *testing.T) {
	s := spec.New(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_sign_in_with_valid_credentials_and_start_a_session",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "api_login_user_1", "api_login_user_1@example.com", "api-login-secret-1")

				res, _ := guestJSON(t, s, http.MethodPost, "/api/login", map[string]string{
					"email":    "api_login_user_1@example.com",
					"password": "api-login-secret-1",
				})
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				sessionRes, body := doJSON(t, s.WrappedHandler(), http.MethodGet, "/api/session", nil, res.Cookies(), "")
				require.Equal(t, http.StatusOK, sessionRes.StatusCode)

				var session apiSessionBody
				require.NoError(t, json.Unmarshal(body, &session))
				require.Equal(t, "api_login_user_1", session.Username)
			},
		},
		{
			name: "should_reject_the_wrong_password",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "api_login_user_2", "api_login_user_2@example.com", "api-login-secret-2")

				res, body := guestJSON(t, s, http.MethodPost, "/api/login", map[string]string{
					"email":    "api_login_user_2@example.com",
					"password": "wrong-password",
				})
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				var errBody struct {
					Error string `json:"error"`
				}
				require.NoError(t, json.Unmarshal(body, &errBody))
				require.Equal(t, "wrong email or password", errBody.Error)
			},
		},
		{
			name: "should_reject_missing_fields_with_per_field_validation",
			fn: func(t *testing.T) {
				res, body := guestJSON(t, s, http.MethodPost, "/api/login", map[string]string{})
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				var errBody struct {
					Fields map[string]string `json:"fields"`
				}
				require.NoError(t, json.Unmarshal(body, &errBody))
				require.Equal(t, "required", errBody.Fields["email"])
				require.Equal(t, "required", errBody.Fields["password"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestAPIRegister(t *testing.T) {
	s := spec.New(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_an_account_and_start_a_session",
			fn: func(t *testing.T) {
				s.CreateInvitationCode(t, "api_register_code_1")

				res, _ := guestJSON(t, s, http.MethodPost, "/api/register", map[string]string{
					"username":              "apiregisteruser1",
					"email":                 "api_register_user_1@example.com",
					"password":              "regpass111",
					"password_confirmation": "regpass111",
					"invitation_code":       "api_register_code_1",
				})
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				sessionRes, body := doJSON(t, s.WrappedHandler(), http.MethodGet, "/api/session", nil, res.Cookies(), "")
				require.Equal(t, http.StatusOK, sessionRes.StatusCode)

				var session apiSessionBody
				require.NoError(t, json.Unmarshal(body, &session))
				require.Equal(t, "apiregisteruser1", session.Username)
			},
		},
		{
			name: "should_reject_a_password_confirmation_mismatch",
			fn: func(t *testing.T) {
				s.CreateInvitationCode(t, "api_register_code_2")

				res, body := guestJSON(t, s, http.MethodPost, "/api/register", map[string]string{
					"username":              "apiregisteruser2",
					"email":                 "api_register_user_2@example.com",
					"password":              "regpass222",
					"password_confirmation": "doesnotmatch",
					"invitation_code":       "api_register_code_2",
				})
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				var errBody struct {
					Error string `json:"error"`
				}
				require.NoError(t, json.Unmarshal(body, &errBody))
				require.Equal(t, "password and password confirmation do not match", errBody.Error)
			},
		},
		{
			name: "should_reject_an_invalid_invitation_code",
			fn: func(t *testing.T) {
				res, body := guestJSON(t, s, http.MethodPost, "/api/register", map[string]string{
					"username":              "apiregisteruser3",
					"email":                 "api_register_user_3@example.com",
					"password":              "regpass333",
					"password_confirmation": "regpass333",
					"invitation_code":       "not-a-real-code",
				})
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				var errBody struct {
					Error string `json:"error"`
				}
				require.NoError(t, json.Unmarshal(body, &errBody))
				require.Equal(t, "invalid invitation code", errBody.Error)
			},
		},
		{
			name: "should_reject_a_duplicate_account",
			fn: func(t *testing.T) {
				s.CreateAuthUser(t, "api_register_user_4", "api_register_user_4@example.com", "api_register_password_4")
				s.CreateInvitationCode(t, "api_register_code_4")

				res, body := guestJSON(t, s, http.MethodPost, "/api/register", map[string]string{
					"username":              "apiregisteruser4b",
					"email":                 "api_register_user_4@example.com",
					"password":              "regpass444",
					"password_confirmation": "regpass444",
					"invitation_code":       "api_register_code_4",
				})
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				var errBody struct {
					Error string `json:"error"`
				}
				require.NoError(t, json.Unmarshal(body, &errBody))
				require.Equal(t, "an account with that username or email already exists", errBody.Error)
			},
		},
		{
			name: "should_reject_missing_fields_with_per_field_validation",
			fn: func(t *testing.T) {
				res, body := guestJSON(t, s, http.MethodPost, "/api/register", map[string]string{})
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				var errBody struct {
					Fields map[string]string `json:"fields"`
				}
				require.NoError(t, json.Unmarshal(body, &errBody))
				require.Equal(t, "required", errBody.Fields["username"])
				require.Equal(t, "required", errBody.Fields["email"])
				require.Equal(t, "required", errBody.Fields["invitation_code"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
