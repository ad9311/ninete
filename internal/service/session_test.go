package service_test

import (
	"context"
	"testing"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSignInUser(t *testing.T) {
	ctx := context.Background()

	store := service.FactoryStore(t)
	defer store.ClosePool()

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_sign_in_the_user",
			func(t *testing.T) {
				user := store.FactoryUser(ctx, t, service.RegistrationParams{})

				session, err := store.SignInUser(ctx, service.SessionParams{
					Email:    user.Email,
					Password: "123456789",
				})

				require.Nil(t, err)
				require.Equal(t, user.Username, session.User.Username)
				require.Equal(t, user.Email, session.User.Email)
				require.NotEmpty(t, session.AccessToken.Value)
				require.NotEmpty(t, session.RefreshToken.Value)
			},
		},
		{
			"should_return_validation_required_error",
			func(t *testing.T) {
				_, err := store.SignInUser(ctx, service.SessionParams{
					Email:    "",
					Password: "",
				})
				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Email]:required")
				require.Contains(t, err.Error(), "[Password]:required")
			},
		},
		{
			"should_return_validation_email_error",
			func(t *testing.T) {
				_, err := store.SignInUser(ctx, service.SessionParams{
					Email:    "wrong@",
					Password: "123456789",
				})
				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Email]:email")
			},
		},
		{
			"should_return_not_found_error",
			func(t *testing.T) {
				_, err := store.SignInUser(ctx, service.SessionParams{
					Email:    "wrong@email.com",
					Password: "123456789",
				})
				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrWrongEmailOrPassword)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestSignOutUser(t *testing.T) {
	ctx := context.Background()

	store := service.FactoryStore(t)
	defer store.ClosePool()

	user := store.FactoryUser(ctx, t, service.RegistrationParams{})
	refreshToken := store.FactorySavedRefreshToken(ctx, t, user.ID)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_sign_out_the_user",
			func(t *testing.T) {
				uuidStr, err := service.UUIDToString(refreshToken.Uuid)
				if err != nil {
					t.Fatalf("failed parsing uuid to string: %v", err)
				}

				err = store.SignOutUser(ctx, uuidStr)

				require.Nil(t, err)
			},
		},
		{
			"should_return_wrong_length_error",
			func(t *testing.T) {
				err := store.SignOutUser(ctx, "1")

				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrInvalidUUIDFormat)
			},
		},
		{
			"should_return_invalid_length_error",
			func(t *testing.T) {
				invalidUUID := "e2edcab4-49cb-48f7-93a7-ea1076fe2-06"
				err := store.SignOutUser(ctx, invalidUUID)

				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrInvalidUUIDLength)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
