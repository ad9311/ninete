package server_test

import (
	"net/http"
	"testing"

	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGetMe(t *testing.T) {
	fs := newFactoryServer(t)

	obj := newFactorySession(t, fs, service.RegistrationParams{})

	res, req := newHTTPTest(factoryHTTP{
		method:      http.MethodGet,
		target:      "/users/me",
		accessToken: obj.AccessToken.Value,
	})
	fs.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	var resBody factoryResponse
	decodeJSONBody(t, res, &resBody)

	var data map[string]service.SafeUser
	dataToStruct(t, resBody.Data, &data)

	require.Equal(t, obj.User.ID, data["user"].ID)
	require.Equal(t, obj.User.Username, data["user"].Username)
	require.Equal(t, obj.User.Email, data["user"].Email)
}
