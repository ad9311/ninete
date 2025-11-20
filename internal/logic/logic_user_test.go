package logic_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestFindUserByEmail(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	params := logic.SignUpParams{
		Username:             "testfinduser1",
		Email:                "testfinduser1@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}

	user := f.User(t, params)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_find_the_user",
			func(t *testing.T) {
				u, err := f.Store.FindUserByEmail(ctx, params.Email)
				require.NoError(t, err)
				require.Equal(t, u.Email, user.Email)
				require.Equal(t, u.Username, user.Username)
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				_, err := f.Store.FindUserByEmail(ctx, "noemail@email.com")
				require.ErrorIs(t, err, logic.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindUser(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	params := logic.SignUpParams{
		Username:             "testfinduser2",
		Email:                "testfinduser2@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}

	user := f.User(t, params)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_find_the_user",
			func(t *testing.T) {
				u, err := f.Store.FindUser(ctx, user.ID)
				require.NoError(t, err)
				require.Equal(t, u.Email, user.Email)
				require.Equal(t, u.Username, user.Username)
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				_, err := f.Store.FindUser(ctx, -1)
				require.ErrorIs(t, err, logic.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
