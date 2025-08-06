package service_test

import (
	"context"
	"testing"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCreateNewRole(t *testing.T) {
	ctx := context.Background()
	store := service.FactoryStore(t)
	defer store.ClosePool()

	roleName := service.FactoryUsername()

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_create_new_role",
			func(t *testing.T) {
				role, err := store.CreateNewRole(ctx, roleName)
				require.Nil(t, err)
				require.Equal(t, roleName, role.Name)
			},
		},
		{
			"should_return_unique_error",
			func(t *testing.T) {
				_, err := store.CreateNewRole(ctx, roleName)
				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrUniqueConstraintViolation)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestFindRoleByName(t *testing.T) {
	ctx := context.Background()
	store := service.FactoryStore(t)
	defer store.ClosePool()

	roleName := "student"
	_ = store.FactoryRole(ctx, t, roleName)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_find_role_by_name",
			func(t *testing.T) {
				role, err := store.FindRoleByName(ctx, roleName)
				require.Nil(t, err)
				require.Equal(t, roleName, role.Name)
			},
		},
		{
			"should_return_not_found_error",
			func(t *testing.T) {
				_, err := store.FindRoleByName(ctx, "non_existing_role")
				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
