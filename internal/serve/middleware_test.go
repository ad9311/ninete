package serve_test

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestJSONMiddleware(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_reject_non_json_post",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodPost, "/auth/sign-in", strings.NewReader(`{"ok":true}`))
				req.Header.Del("Content-Type")

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnsupportedMediaType, res.Code)
			},
		},
		{
			"should_allow_json_post",
			func(t *testing.T) {
				params := logic.SessionParams{
					Email:    "signin-json@example.com",
					Password: "123456789",
				}
				userParams := logic.SignUpParams{
					Username:             "signinjson",
					Email:                params.Email,
					Password:             params.Password,
					PasswordConfirmation: params.Password,
				}
				f.User(t, userParams)

				body := testhelper.MarshalPayload(t, params)
				res, req := f.NewRequest(ctx, http.MethodPost, "/auth/sign-in", body)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusCreated, res.Code)
			},
		},
		{
			"should_allow_non_post_without_content_type",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodGet, "/healthz", nil)
				req.Header.Del("Content-Type")

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusNoContent, res.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestCORSMiddleware(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	require.NotEmpty(t, origins)
	allowedOrigin := strings.TrimSpace(origins[0])

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_allow_preflight_for_allowed_origin",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodOptions, "/auth/sign-in", nil)
				req.Header.Set("Origin", allowedOrigin)
				req.Header.Set("Access-Control-Request-Method", "POST")
				req.Header.Set("Access-Control-Request-Headers", "Content-Type")

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusNoContent, res.Code)
				require.Equal(t, allowedOrigin, res.Header().Get("Access-Control-Allow-Origin"))
				require.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", res.Header().Get("Access-Control-Allow-Methods"))
				require.Equal(t, "Content-Type", res.Header().Get("Access-Control-Allow-Headers"))
				require.Equal(t, "1800", res.Header().Get("Access-Control-Max-Age"))
			},
		},
		{
			"should_reject_preflight_for_disallowed_origin",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodOptions, "/auth/sign-in", nil)
				req.Header.Set("Origin", "https://not-allowed.com")
				req.Header.Set("Access-Control-Request-Method", "POST")

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusForbidden, res.Code)
			},
		},
		{
			"should_reject_simple_request_for_disallowed_origin",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodGet, "/readyz", nil)
				req.Header.Set("Origin", "https://not-allowed.com")

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusOK, res.Code)
				require.Empty(t, res.Header().Get("Access-Control-Allow-Origin"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestNotFoundAndMethodNotAllowedHandlers(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_return_not_found",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodGet, "/unknown", nil)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusNotFound, res.Code)
			},
		},
		{
			"should_return_method_not_allowed",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodPost, "/readyz", nil)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusMethodNotAllowed, res.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestAuthMiddleware(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	params := logic.SignUpParams{
		Username:             "authmiddleware",
		Email:                "authmiddleware@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, params)

	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	unknownToken, err := f.Store.NewAccessToken(-1)
	require.NoError(t, err)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_reject_missing_bearer",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodGet, "/users/me", nil)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)
			},
		},
		{
			"should_reject_invalid_token",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodGet, "/users/me", nil)
				req.Header.Set("Authorization", "Bearer invalid")

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)
			},
		},
		{
			"should_allow_valid_token",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodGet, "/users/me", nil)
				req.Header.Set("Authorization", "Bearer "+token.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusOK, res.Code)

				var payload testhelper.Response[repo.SafeUser]
				require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
				require.Equal(t, user.Email, payload.Data.Email)
			},
		},
		{
			"should_reject_unknown_user",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodGet, "/users/me", nil)
				req.Header.Set("Authorization", "Bearer "+unknownToken.Value)

				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
