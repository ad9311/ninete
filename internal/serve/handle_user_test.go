package serve_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestGetMe(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	params := logic.SignUpParams{
		Username:             "metest",
		Email:                "me@example.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}
	user := f.User(t, params)

	signInRes := f.SignInUser(t, ctx, logic.SessionParams{
		Email:    params.Email,
		Password: params.Password,
	})

	res, req := f.NewRequest(ctx, http.MethodGet, "/users/me", nil)
	testhelper.SetAuthHeader(req, signInRes.Data.AccessToken.Value)
	req = req.WithContext(context.WithValue(req.Context(), prog.KeyCurrentUser, &user))

	f.Server.Router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	var payload testhelper.Response[repo.SafeUser]
	testhelper.UnmarshalPayload(t, res, &payload)
	require.Equal(t, user.ID, payload.Data.ID)
	require.Equal(t, user.Email, payload.Data.Email)
	require.Equal(t, user.Username, payload.Data.Username)
}
