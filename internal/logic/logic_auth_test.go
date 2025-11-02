package logic_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/stretchr/testify/require"
)

func TestSignUpUser(t *testing.T) {
	ctx := t.Context()

	params := logic.SignUpParams{
		Username:             "testsignup",
		Email:                "testsignup@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}

	store := newTestStore(t)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_sign_up_user",
			func(t *testing.T) {
				user, err := store.SignUpUser(ctx, params)
				require.NoError(t, err)
				require.Equal(t, user.Email, params.Email)
				require.Equal(t, user.Username, params.Username)
				require.Positive(t, user.ID)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
