package serve_test

import (
	"net/http"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
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
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_register_user",
			func(t *testing.T) {
				body := testhelper.MarshalPayload(t, params)
				res, req := f.NewRequest(ctx, http.MethodPost, target, body)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusCreated, res.Code)

				var payload testhelper.Response[repo.SafeUser]
				testhelper.UnmarshalPayload(t, res, &payload)
				require.Equal(t, payload.Data.Username, params.Username)
				require.Equal(t, payload.Data.Email, params.Email)
				require.Positive(t, payload.Data.ID)
			},
		},
		{
			"should_fail_existing_user",
			func(t *testing.T) {
				body := testhelper.MarshalPayload(t, params)
				res, req := f.NewRequest(ctx, http.MethodPost, target, body)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var payload testhelper.FailedResponse
				testhelper.UnmarshalPayload(t, res, &payload)
				require.Nil(t, payload.Data)
				require.Equal(t, payload.Error, "UNIQUE constraint failed: users.email")
			},
		},
		{
			"should_fail_validations",
			func(t *testing.T) {
				body := testhelper.MarshalPayload(t, logic.SignUpParams{})
				res, req := f.NewRequest(ctx, http.MethodPost, target, body)
				f.Server.Router.ServeHTTP(res, req)

				require.Equal(t, http.StatusBadRequest, res.Code)

				var payload testhelper.FailedResponse
				testhelper.UnmarshalPayload(t, res, &payload)
				require.Nil(t, payload.Data)
				require.Equal(
					t,
					payload.Error,
					"validation failed: [Username:required],[Email:email],[Password:min],[PasswordConfirmation:min]",
				)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
