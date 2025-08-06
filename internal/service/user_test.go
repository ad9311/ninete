package service_test

import (
	"context"
	"testing"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestFindUserByID(t *testing.T) {
	ctx := context.Background()
	store := service.FactoryStore(t)
	defer store.ClosePool()

	user := store.FactoryUser(ctx, t, service.RegistrationParams{})

	cases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "should_find_user",
			testFunc: func(t *testing.T) {
				foundUser, err := store.FindUserByID(ctx, user.ID)
				require.Nil(t, err)
				require.Equal(t, user.ID, foundUser.ID)
			},
		},
		{
			name: "should_return_not_found_error",
			testFunc: func(t *testing.T) {
				_, err := store.FindUserByID(ctx, -1)
				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestFindUserByEmail(t *testing.T) {
	ctx := context.Background()
	store := service.FactoryStore(t)
	defer store.ClosePool()

	user := store.FactoryUser(ctx, t, service.RegistrationParams{})

	cases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "should_find_user",
			testFunc: func(t *testing.T) {
				foundUser, err := store.FindUserByEmail(ctx, user.Email)
				require.Nil(t, err)
				require.Equal(t, user.Email, foundUser.Email)
			},
		},
		{
			name: "should_return_not_found_error",
			testFunc: func(t *testing.T) {
				_, err := store.FindUserByEmail(ctx, "notfound@example.com")
				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestGetUserRoles(t *testing.T) {
	ctx := context.Background()
	store := service.FactoryStore(t)
	defer store.ClosePool()

	user := store.FactoryUser(ctx, t, service.RegistrationParams{})
	for range 5 {
		role := store.FactoryRole(ctx, t, "")
		_ = store.FactoryUserRole(ctx, t, user.ID, role.Name)
	}

	cases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			"should_get_user_roles",
			func(t *testing.T) {
				roles, err := store.GetUserRoles(ctx, user.ID)
				require.Nil(t, err)
				require.Len(t, roles, 5)
			},
		},
		{
			"should_not_get_roles",
			func(t *testing.T) {
				user := store.FactoryUser(ctx, t, service.RegistrationParams{})
				roles, err := store.GetUserRoles(ctx, user.ID)
				require.Nil(t, err)
				require.Len(t, roles, 0)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
