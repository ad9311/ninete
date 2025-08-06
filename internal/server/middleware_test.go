package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestJSONMiddleware(t *testing.T) {
	fs := newFactoryServer(t)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_set_content_type_header",
			func(t *testing.T) {
				res, req := newHTTPTest(factoryHTTP{
					method: http.MethodGet,
					target: "/healthz",
				})
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusNoContent, res.Code)
				require.Equal(t, "application/json", res.Header().Get("Content-Type"))
			},
		},
		{
			"should_enforce_content_type_header",
			func(t *testing.T) {
				methods := []string{
					http.MethodPost,
					http.MethodPatch,
					http.MethodPut,
					http.MethodDelete,
				}
				for _, m := range methods {
					res := httptest.NewRecorder()
					req := httptest.NewRequest(m, "/auth/sign-in", nil)
					fs.router.ServeHTTP(res, req)

					require.Equal(t, http.StatusUnsupportedMediaType, res.Code)

					var resBody factoryErrorResponse
					decodeJSONBody(t, res, &resBody)

					require.Equal(t, errs.ErrUnsupportedMediaType.Error(), resBody.Error)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestNotFoundHandler(t *testing.T) {
	fs := newFactoryServer(t)

	res, req := newHTTPTest(factoryHTTP{
		method: http.MethodGet,
		target: "/unknown-route",
	})
	fs.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestMethodNotAllowedHandler(t *testing.T) {
	fs := newFactoryServer(t)

	res, req := newHTTPTest(factoryHTTP{
		method: http.MethodPost,
		target: "/healthz",
	})
	fs.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusMethodNotAllowed, res.Code)
}

func TestAuthMiddleware(t *testing.T) {
	fs := newFactoryServer(t)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_authenticate_a_user",
			func(t *testing.T) {
				sess := newFactorySession(t, fs, service.RegistrationParams{})

				res, req := newHTTPTest(factoryHTTP{
					method:      http.MethodGet,
					target:      "/users/me",
					accessToken: sess.AccessToken.Value,
				})
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusOK, res.Code)
			},
		},
		{
			"should_reject_an_authenticated_user",
			func(t *testing.T) {
				res, req := newHTTPTest(factoryHTTP{
					method: http.MethodGet,
					target: "/users/me",
				})
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)

				var resBody factoryErrorResponse
				decodeJSONBody(t, res, &resBody)

				require.Equal(t, errs.ErrInvalidAuthHeader.Error(), resBody.Error)
			},
		},
		{
			"should_reject_when_invalid_token",
			func(t *testing.T) {
				res, req := newHTTPTest(factoryHTTP{
					method:      http.MethodGet,
					target:      "/users/me",
					accessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
				})
				fs.router.ServeHTTP(res, req)

				require.Equal(t, http.StatusUnauthorized, res.Code)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
