package serve_test

import (
	"net/http"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestPostSignUp(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	target := "/auth/sign-up"
	params := logic.SignUpParams{
		Username:             "testsignsup",
		Email:                "testsignup@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_register_user",
			func(t *testing.T) {
				body := testhelper.MarshalPayload(t, params)
				res, req := f.NewRequest(ctx, http.MethodPost, target, body)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusCreated, res.Code)

				var resBody testhelper.Response[repo.SafeUser]
				testhelper.UnmarshalBody(t, res, &resBody)
				require.Equal(t, resBody.Data.Username, params.Username)
				require.Equal(t, resBody.Data.Email, params.Email)
				require.Positive(t, resBody.Data.ID)
			},
		},
		{
			"should_fail_existing_user",
			func(t *testing.T) {
				body := testhelper.MarshalPayload(t, params)
				res, req := f.NewRequest(ctx, http.MethodPost, target, body)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &resBody)
				require.Nil(t, resBody.Data)
				require.Equal(t, resBody.Error, "UNIQUE constraint failed: users.email")
			},
		},
		{
			"should_fail_validations",
			func(t *testing.T) {
				body := testhelper.MarshalPayload(t, logic.SignUpParams{})
				res, req := f.NewRequest(ctx, http.MethodPost, target, body)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &resBody)
				require.Nil(t, resBody.Data)
				require.Contains(t, resBody.Error, logic.ErrValidationFailed.Error())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostSignIn(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	target := "/auth/sign-in"

	signUpParams := logic.SignUpParams{
		Username:             "signintest",
		Email:                "signin@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	f.User(t, signUpParams)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_sign_in_user",
			func(t *testing.T) {
				params := logic.SessionParams{
					Email:    signUpParams.Email,
					Password: signUpParams.Password,
				}
				body := testhelper.MarshalPayload(t, params)
				res, req := f.NewRequest(ctx, http.MethodPost, target, body)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusCreated, res.Code)

				var resBody testhelper.Response[serve.SessionResponse]
				testhelper.UnmarshalBody(t, res, &resBody)
				require.Equal(t, signUpParams.Email, resBody.Data.User.Email)
				require.NotEmpty(t, resBody.Data.AccessToken.Value)

				resp := res.Result()
				var found bool
				for _, cookie := range resp.Cookies() {
					if cookie.Name == "refresh_token" {
						found = true
						require.NotEmpty(t, cookie.Value)
						require.Equal(t, "/auth", cookie.Path)
						require.True(t, cookie.HttpOnly)

						break
					}
				}
				require.True(t, found)
			},
		},
		{
			"should_fail_invalid_credentials",
			func(t *testing.T) {
				params := logic.SessionParams{
					Email:    signUpParams.Email,
					Password: "wrong-password",
				}
				body := testhelper.MarshalPayload(t, params)
				res, req := f.NewRequest(ctx, http.MethodPost, target, body)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &resBody)
				require.Nil(t, resBody.Data)
				require.Equal(t, logic.ErrWrongEmailOrPassword.Error(), resBody.Error)
			},
		},
		{
			"should_fail_validations",
			func(t *testing.T) {
				body := testhelper.MarshalPayload(t, logic.SessionParams{})
				res, req := f.NewRequest(ctx, http.MethodPost, target, body)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &resBody)
				require.Nil(t, resBody.Data)
				require.Contains(t, resBody.Error, logic.ErrValidationFailed.Error())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteSignOut(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	target := "/auth/sign-out"

	signUpParams := logic.SignUpParams{
		Username:             "signouttest",
		Email:                "signout@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, signUpParams)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_delete_refresh_token",
			func(t *testing.T) {
				token := f.RefreshToken(t, user.ID)
				res, req := f.NewRequest(ctx, http.MethodDelete, target, nil)
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: token.Value,
					Path:  "/auth",
				})
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusNoContent, res.Code)

				resp := res.Result()
				var found bool
				for _, cookie := range resp.Cookies() {
					if cookie.Name == "refresh_token" {
						found = true
						require.Empty(t, cookie.Value)
						require.Equal(t, "/auth", cookie.Path)
						require.True(t, cookie.HttpOnly)

						break
					}
				}
				require.True(t, found)

				_, err := f.Store.FindRefreshToken(ctx, token.Value)
				require.ErrorIs(t, err, logic.ErrNotFound)
			},
		},
		{
			"should_return_no_content_without_cookie",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodDelete, target, nil)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusNoContent, res.Code)

				resp := res.Result()
				var found bool
				for _, cookie := range resp.Cookies() {
					if cookie.Name == "refresh_token" {
						found = true
						require.Empty(t, cookie.Value)
						require.Equal(t, "/auth", cookie.Path)
						require.True(t, cookie.HttpOnly)

						break
					}
				}
				require.True(t, found)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPostRefresh(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	target := "/auth/refresh"

	signUpParams := logic.SignUpParams{
		Username:             "refreshtest",
		Email:                "refresh@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, signUpParams)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_return_access_token",
			func(t *testing.T) {
				token := f.RefreshToken(t, user.ID)
				res, req := f.NewRequest(ctx, http.MethodPost, target, nil)
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: token.Value,
					Path:  "/auth",
				})
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusOK, res.Code)

				var resBody testhelper.Response[serve.SessionResponse]
				testhelper.UnmarshalBody(t, res, &resBody)
				require.NotEmpty(t, resBody.Data.AccessToken.Value)
				require.Greater(t, resBody.Data.AccessToken.ExpiresAt, resBody.Data.AccessToken.IssuedAt)
				require.Positive(t, resBody.Data.User.ID)
			},
		},
		{
			"should_fail_without_cookie",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodPost, target, nil)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)

				var resBody testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &resBody)
				require.Nil(t, resBody.Data)
				require.Contains(t, resBody.Error, serve.ErrInvalidAuthCreds.Error())
				require.Contains(t, resBody.Error, "refresh cookie not found")
			},
		},
		{
			"should_fail_with_invalid_token",
			func(t *testing.T) {
				res, req := f.NewRequest(ctx, http.MethodPost, target, nil)
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "invalid-token",
					Path:  "/auth",
				})
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)

				var resBody testhelper.FailedResponse
				testhelper.UnmarshalBody(t, res, &resBody)
				require.Nil(t, resBody.Data)
				require.Contains(t, resBody.Error, serve.ErrInvalidAuthCreds.Error())
				require.Contains(t, resBody.Error, logic.ErrNotFound.Error())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
