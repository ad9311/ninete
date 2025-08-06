package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/server"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPostSignIn(t *testing.T) {
	fs := newFactoryServer(t)

	username := service.FactoryUsername()
	body := newRequestBody(t, service.RegistrationParams{
		Username:             username,
		Email:                username + "@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	user := signUpUser(t, fs, bytes.NewReader(body))

	reqFunc := func(body []byte) (*httptest.ResponseRecorder, *http.Request) {
		res, req := newHTTPTest(factoryHTTP{
			method: http.MethodPost,
			target: "/auth/sign-in",
			body:   bytes.NewReader(body),
		})

		return res, req
	}

	cases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			"should_sign_in_user",
			func(t *testing.T) {
				body := newRequestBody(t, service.SessionParams{
					Email:    user.Email,
					Password: "123456789",
				})

				res, req := reqFunc(body)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusCreated, res.Code)

				var resBody factoryResponse
				decodeJSONBody(t, res, &resBody)

				var data server.SessionResponse
				dataToStruct(t, resBody.Data, &data)

				require.Equal(t, user.ID, data.User.ID)
				require.Equal(t, user.Username, data.User.Username)
				require.Equal(t, user.Email, data.User.Email)
				require.NotEmpty(t, data.AccessToken.Value)
				require.NotEmpty(t, data.AccessToken.IssuedAt)
				require.NotEmpty(t, data.AccessToken.ExpiresAt)
			},
		},
		{
			"should_return_validation_error",
			func(t *testing.T) {
				body := newRequestBody(t, service.SessionParams{
					Email:    "",
					Password: "",
				})

				res, req := reqFunc(body)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Contains(t, resBody.Error, errs.ErrValidationFailed.Error())
				require.Contains(t, resBody.Error, "[Email]:required")
				require.Contains(t, resBody.Error, "[Password]:required")
			},
		},
		{
			"should_return_email_format_validation_error",
			func(t *testing.T) {
				body := newRequestBody(t, service.SessionParams{
					Email:    "wrong_email@",
					Password: "123456789",
				})

				res, req := reqFunc(body)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Contains(t, resBody.Error, errs.ErrValidationFailed.Error())
				require.Contains(t, resBody.Error, "[Email]:email")
			},
		},
		{
			"should_return_error_when_user_not_found",
			func(t *testing.T) {
				body := newRequestBody(t, service.SessionParams{
					Email:    "wrong_email@email.com",
					Password: "123456789",
				})

				res, req := reqFunc(body)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Equal(t, errs.ErrWrongEmailOrPassword.Error(), resBody.Error)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestPostRefresh(t *testing.T) {
	fs := newFactoryServer(t)

	sess := newFactorySession(t, fs, service.RegistrationParams{})

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_refresh_the_acces_token",
			func(t *testing.T) {
				res, req := newHTTPTest(factoryHTTP{
					method: http.MethodPost,
					target: "/auth/refresh",
				})
				req.AddCookie(sess.RefreshTokenCookie)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusCreated, res.Code)

				var resBody factoryResponse
				decodeJSONBody(t, res, &resBody)

				var data server.SessionResponse
				dataToStruct(t, resBody.Data, &data)

				require.Equal(t, sess.User.ID, data.User.ID)
				require.Equal(t, sess.User.Username, data.User.Username)
				require.Equal(t, sess.User.Email, data.User.Email)
				require.NotEmpty(t, data.AccessToken.Value)
				require.NotEmpty(t, data.AccessToken.IssuedAt)
				require.NotEmpty(t, data.AccessToken.ExpiresAt)
			},
		},
		{
			"should_return_unauthorized_when_no_cookie",
			func(t *testing.T) {
				res, req := newHTTPTest(factoryHTTP{
					method: http.MethodPost,
					target: "/auth/refresh",
				})
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Equal(t, "INVALID_AUTHENTICATION_CREDENTIALS", resBody.Code)
				require.Equal(t, errs.ErrRefreshTokenNotFound.Error(), resBody.Error)
			},
		},
		{
			"should_return_unauthorized_when_invalid_cookie",
			func(t *testing.T) {
				res, req := newHTTPTest(factoryHTTP{
					method: http.MethodPost,
					target: "/auth/refresh",
				})
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "invalid cookie value",
				})
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Equal(t, "ERROR", resBody.Code)
				require.Equal(t, errs.ErrInvalidUUIDFormat.Error(), resBody.Error)
			},
		},
		{
			"should_return_unauthorized_when_cookie_not_found",
			func(t *testing.T) {
				res, req := newHTTPTest(factoryHTTP{
					method: http.MethodPost,
					target: "/auth/refresh",
				})
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "11111111-1111-1111-1111-111111111111",
				})
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Equal(t, "ERROR", resBody.Code)
				require.Equal(t, errs.ErrNotFound.Error(), resBody.Error)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestDeleteSignOut(t *testing.T) {
	fs := newFactoryServer(t)

	sess := newFactorySession(t, fs, service.RegistrationParams{})

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_sign_out_the_user",
			func(t *testing.T) {
				res, req := newHTTPTest(factoryHTTP{
					method:      http.MethodDelete,
					target:      "/auth/sign-out",
					accessToken: sess.AccessToken.Value,
				})
				req.AddCookie(sess.RefreshTokenCookie)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusNoContent, res.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
