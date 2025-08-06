package service_test

import (
	"context"
	"testing"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestSaveRefreshToken(t *testing.T) {
	ctx := context.Background()

	store := service.FactoryStore(t)
	defer store.ClosePool()

	user := store.FactoryUser(ctx, t, service.RegistrationParams{})

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_save_the_token",
			func(t *testing.T) {
				_, err := store.SaveRefreshToken(ctx, user.ID)
				require.Nil(t, err)
			},
		},
		{
			"should_return_error",
			func(t *testing.T) {
				_, err := store.SaveRefreshToken(ctx, -1)
				require.NotNil(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestFindRefreshTokenByUUID(t *testing.T) {
	ctx := context.Background()

	store := service.FactoryStore(t)
	defer store.ClosePool()

	user := store.FactoryUser(ctx, t, service.RegistrationParams{})
	refreshToken := store.FactorySavedRefreshToken(ctx, t, user.ID)
	tokenStr, err := service.UUIDToString(refreshToken.Uuid)
	if err != nil {
		t.Fatalf("failed to parse factory refresh token: %v", err)
	}

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_save_the_token",
			func(t *testing.T) {
				foundToken, err := store.FindRefreshTokenByUUID(ctx, tokenStr)

				require.Nil(t, err)
				require.Equal(t, refreshToken.Uuid, foundToken.Uuid)
			},
		},
		{
			"should_return_not_found_error",
			func(t *testing.T) {
				invalidToken := "11111111-1111-1111-1111-111111111111"
				_, err := store.FindRefreshTokenByUUID(ctx, invalidToken)

				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestDeleteExpiredRefreshTokens(t *testing.T) {
	ctx := context.Background()

	store := service.FactoryStore(t)
	defer store.ClosePool()

	user := store.FactoryUser(ctx, t, service.RegistrationParams{})

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_delete_the_token",
			func(t *testing.T) {
				var uuids []pgtype.UUID
				for range 5 {
					refreshToken := store.FactoryExpiredToken(ctx, t, user.ID)
					uuids = append(uuids, refreshToken.Uuid)
				}

				count, err := store.DeleteExpieredRefreshTokens(ctx)
				require.Nil(t, err)
				require.Equal(t, int64(5), count)

				for _, u := range uuids {
					tokenStr, err := service.UUIDToString(u)
					if err != nil {
						t.Fatalf("failed to parse factory refresh token: %v", err)
					}
					_, err = store.FindRefreshTokenByUUID(ctx, tokenStr)
					require.NotNil(t, err)
					require.ErrorIs(t, err, errs.ErrNotFound)
				}
			},
		},
		{
			"should_not_delete_non_expired_tokens",
			func(t *testing.T) {
				_ = store.FactorySavedRefreshToken(ctx, t, user.ID)
				count, err := store.DeleteExpieredRefreshTokens(ctx)
				require.Nil(t, err)
				require.Equal(t, int64(0), count)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
