package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPostSignUp(t *testing.T) {
	fs := newFactoryServer(t)

	params := service.RegistrationParams{
		Username:             "testing_sign_up",
		Email:                "testing_sign_up@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}

	reqFunc := func(body []byte) (*httptest.ResponseRecorder, *http.Request) {
		return newHTTPTest(factoryHTTP{
			method: http.MethodPost,
			target: "/auth/sign-up",
			body:   bytes.NewReader(body),
		})
	}

	cases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			"should_register_the_user",
			func(t *testing.T) {
				body := newRequestBody(t, params)
				res, req := reqFunc(body)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusCreated, res.Code)

				var resBody factoryResponse
				decodeJSONBody(t, res, &resBody)

				var data service.SafeUser
				dataToStruct(t, resBody.Data, &data)

				require.Equal(t, params.Username, data.Username)
				require.Equal(t, params.Email, data.Email)
			},
		},
		{
			"should_return_validation_required_error",
			func(t *testing.T) {
				body := newRequestBody(t, service.RegistrationParams{})
				res, req := reqFunc(body)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Equal(t, "INVALID_FORM", resBody.Code)
				require.Contains(t, resBody.Error, errs.ErrValidationFailed.Error())
				require.Contains(t, resBody.Error, "[Username]:required")
				require.Contains(t, resBody.Error, "[Email]:required")
				require.Contains(t, resBody.Error, "[Password]:required")
				require.Contains(t, resBody.Error, "[PasswordConfirmation]:required")
			},
		},
		{
			"should_return_validation_min_error",
			func(t *testing.T) {
				body := newRequestBody(t, service.RegistrationParams{
					Username:             "12",
					Email:                params.Username,
					Password:             "1234567",
					PasswordConfirmation: "1234567",
				})
				res, req := reqFunc(body)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Equal(t, "INVALID_FORM", resBody.Code)
				require.Contains(t, resBody.Error, errs.ErrValidationFailed.Error())
				require.Contains(t, resBody.Error, "[Username]:min")
				require.Contains(t, resBody.Error, "[Password]:min")
				require.Contains(t, resBody.Error, "[PasswordConfirmation]:min")
			},
		},
		{
			"should_return_validation_max_error",
			func(t *testing.T) {
				body := newRequestBody(t, service.RegistrationParams{
					Username:             "11111111111111111111111111111111111111111111111",
					Email:                params.Email,
					Password:             "11111111111111111111111111111111111111111111111",
					PasswordConfirmation: "11111111111111111111111111111111111111111111111",
				})
				res, req := reqFunc(body)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Equal(t, "INVALID_FORM", resBody.Code)
				require.Contains(t, resBody.Error, errs.ErrValidationFailed.Error())
				require.Contains(t, resBody.Error, "[Username]:max")
				require.Contains(t, resBody.Error, "[Password]:max")
				require.Contains(t, resBody.Error, "[PasswordConfirmation]:max")
			},
		},
		{
			"should_return_validation_email_error",
			func(t *testing.T) {
				body := newRequestBody(t, service.RegistrationParams{
					Username:             params.Username,
					Email:                params.Username,
					Password:             params.Password,
					PasswordConfirmation: params.PasswordConfirmation,
				})
				res, req := reqFunc(body)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Equal(t, "INVALID_FORM", resBody.Code)
				require.Contains(t, resBody.Error, errs.ErrValidationFailed.Error())
				require.Contains(t, resBody.Error, "[Email]:email")
			},
		},
		{
			"should_return_unmatched_passwords_error",
			func(t *testing.T) {
				body := newRequestBody(t, service.RegistrationParams{
					Username:             params.Username,
					Email:                params.Username,
					Password:             params.Password,
					PasswordConfirmation: "123ishaisdhiu34",
				})
				res, req := reqFunc(body)
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Equal(t, "INVALID_FORM", resBody.Code)
				require.Equal(t, errs.ErrUnmatchedPasswords.Error(), resBody.Error)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
