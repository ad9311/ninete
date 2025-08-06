package service_test

import (
	"context"
	"testing"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRegisterUser(t *testing.T) {
	ctx := context.Background()

	store := service.FactoryStore(t)
	defer store.ClosePool()

	username := service.FactoryUsername()
	validParams := service.RegistrationParams{
		Username:             username,
		Email:                username + "@email.com",
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
				user, err := store.RegisterUser(ctx, validParams)

				require.Nil(t, err)
				require.Positive(t, user.ID)
				require.Equal(t, validParams.Username, user.Username)
				require.Equal(t, validParams.Email, user.Email)
			},
		},
		{
			"should_return_unmatched_passwords_error",
			func(t *testing.T) {
				params := validParams
				params.Password = ""

				_, err := store.RegisterUser(ctx, params)
				require.NotNil(t, err)
				require.Contains(t, err.Error(), errs.ErrUnmatchedPasswords.Error())
			},
		},
		{
			"should_return_validation_required_error",
			func(t *testing.T) {
				params := validParams
				params.Username = ""
				params.Email = ""

				_, err := store.RegisterUser(ctx, params)
				require.NotNil(t, err)
				require.Contains(t, err.Error(), "[Email]:required")
				require.Contains(t, err.Error(), "[Username]:required")
			},
		},
		{
			"should_return_validation_email_error",
			func(t *testing.T) {
				params := validParams
				params.Email = "wrong_email@"

				_, err := store.RegisterUser(ctx, params)
				require.NotNil(t, err)
				require.Contains(t, err.Error(), "[Email]:email")
			},
		},
		{
			"should_return_validation_min_error",
			func(t *testing.T) {
				params := validParams
				params.Username = "wr"
				params.Password = "1"
				params.PasswordConfirmation = "1"

				_, err := store.RegisterUser(ctx, params)
				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Username]:min")
				require.Contains(t, err.Error(), "[Password]:min")
				require.Contains(t, err.Error(), "[PasswordConfirmation]:min")
			},
		},
		{
			"should_return_validation_max_error",
			func(t *testing.T) {
				str := "123456789101112312312312313"

				params := validParams
				params.Username = str
				params.Password = str
				params.PasswordConfirmation = str

				_, err := store.RegisterUser(ctx, params)
				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Username]:max")
				require.Contains(t, err.Error(), "[Password]:max")
				require.Contains(t, err.Error(), "[PasswordConfirmation]:max")
			},
		},
		{
			"should_return_unique_error",
			func(t *testing.T) {
				_, err := store.RegisterUser(ctx, validParams)
				require.NotNil(t, err)
				require.Contains(t, err.Error(), errs.ErrUniqueConstraintViolation.Error())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
