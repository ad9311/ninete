package logic_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/testhelper"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestSignUpUser(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	params := logic.SignUpParams{
		Username:             "testsignup",
		Email:                "testsignup@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_sign_up_user",
			func(t *testing.T) {
				user, err := f.Store.SignUpUser(ctx, params)
				require.NoError(t, err)
				require.Equal(t, user.Email, params.Email)
				require.Equal(t, user.Username, params.Username)
				require.Positive(t, user.ID)
			},
		},
		{
			"should_fail_when_password_do_not_match",
			func(t *testing.T) {
				fParams := params
				fParams.PasswordConfirmation = ""
				_, err := f.Store.SignUpUser(ctx, fParams)
				require.ErrorIs(t, err, logic.ErrWithPasswords)
				require.Equal(t, err.Error(), logic.ErrWithPasswords.Error()+", they do not match")
			},
		},
		{
			"should_fail_username_validation",
			func(t *testing.T) {
				fParams := params

				fParams.Username = "$@#^%#@"
				_, err := f.Store.SignUpUser(ctx, fParams)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Username:alphanumunicode]")

				fParams.Username = ""
				_, err = f.Store.SignUpUser(ctx, fParams)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Username:required]")

				fParams.Username = "a"
				_, err = f.Store.SignUpUser(ctx, fParams)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Username:min]")

				fParams.Username = "akasjdhakshdakshdkueyhuehjdd"
				_, err = f.Store.SignUpUser(ctx, fParams)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Username:max]")
			},
		},
		{
			"should_fail_email_validation",
			func(t *testing.T) {
				fParams := params

				fParams.Email = "invalid-email"
				_, err := f.Store.SignUpUser(ctx, fParams)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Email:email]")

				fParams.Email = ""
				_, err = f.Store.SignUpUser(ctx, fParams)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Email:email]")
			},
		},
		{
			"should_fail_password_validation",
			func(t *testing.T) {
				fParams := params

				fParams.Password = "short"
				fParams.PasswordConfirmation = "short"
				_, err := f.Store.SignUpUser(ctx, fParams)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Password:min]")
				require.Contains(t, err.Error(), "[PasswordConfirmation:min]")

				fParams.Password = "123456789012345678901"
				fParams.PasswordConfirmation = "123456789012345678901"
				_, err = f.Store.SignUpUser(ctx, fParams)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Password:max]")
				require.Contains(t, err.Error(), "[PasswordConfirmation:max]")
			},
		},
		{
			"should_fail_unique_email",
			func(t *testing.T) {
				fParams := params
				_, err := f.Store.SignUpUser(ctx, fParams)

				testUniqueConstraint(t, err, "email")
			},
		},
		{
			"should_fail_unique_username",
			func(t *testing.T) {
				fParams := params
				fParams.Email = "testsignup2@email.com"
				_, err := f.Store.SignUpUser(ctx, fParams)

				testUniqueConstraint(t, err, "username")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func testUniqueConstraint(t *testing.T, err error, field string) {
	require.Error(t, err)

	var sqlErr sqlite3.Error
	require.ErrorAs(t, err, &sqlErr)
	require.Equal(t, sqlite3.ErrConstraint, sqlErr.Code)
	require.Equal(t, sqlite3.ErrConstraintUnique, sqlErr.ExtendedCode)
	require.Contains(t, err.Error(), "users."+field)
}
