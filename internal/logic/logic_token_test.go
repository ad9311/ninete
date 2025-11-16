package logic_test

import (
	"strconv"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestFindRefreshToken(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "refreshtoken",
		Email:                "refreshtoken@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	refreshToken := f.RefreshToken(t, user.ID)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_find_refresh_token",
			func(t *testing.T) {
				rf, err := f.Store.FindRefreshToken(ctx, refreshToken.Value)
				require.NoError(t, err)
				require.Equal(t, rf.UserID, user.ID)
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				_, err := f.Store.FindRefreshToken(ctx, "")
				require.ErrorIs(t, err, logic.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestNewRefreshToken(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "refreshtoken2",
		Email:                "refreshtoken2@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_create_refresh_token",
			func(t *testing.T) {
				rf, err := f.Store.NewRefreshToken(ctx, user.ID)
				require.NoError(t, err)
				require.NotEmpty(t, rf.Value)
			},
		},
		{
			"should_failed_unexisting_user",
			func(t *testing.T) {
				_, err := f.Store.NewRefreshToken(ctx, -1)
				require.Error(t, err)
				require.Contains(t, err.Error(), "FOREIGN KEY constraint failed")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestNewAccessToken(t *testing.T) {
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "accesstoken",
		Email:                "accesstoken@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_create_access_token",
			func(t *testing.T) {
				token, err := f.Store.NewAccessToken(user.ID)
				require.NoError(t, err)
				require.NotEmpty(t, token.Value)
				require.Greater(t, token.ExpiresAt, token.IssuedAt)

				claims, err := f.Store.ParseAndValidateJWT(token.Value)
				require.NoError(t, err)
				require.Equal(t, strconv.Itoa(user.ID), claims["sub"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestParseAndValidateJWT(t *testing.T) {
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "parsejwt",
		Email:                "parsejwt@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})

	token, err := f.Store.NewAccessToken(user.ID)
	require.NoError(t, err)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "should_parse_valid_token",
			testFunc: func(t *testing.T) {
				claims, err := f.Store.ParseAndValidateJWT(token.Value)
				require.NoError(t, err)
				require.Equal(t, strconv.Itoa(user.ID), claims["sub"])
			},
		},
		{
			name: "should_fail_invalid_token",
			testFunc: func(t *testing.T) {
				_, err := f.Store.ParseAndValidateJWT(token.Value + "tampered")
				require.Error(t, err)
				require.Contains(t, err.Error(), "token signature is invalid: signature is invalid")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
