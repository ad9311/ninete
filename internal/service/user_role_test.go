package service_test

import (
	"context"
	"testing"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAddRoleToUser(t *testing.T) {
	ctx := context.Background()
	store := service.FactoryStore(t)

	user := store.FactoryUser(ctx, t, service.RegistrationParams{})
	role := store.FactoryRole(ctx, t, "")

	cases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			"should_add_role_to_user",
			func(t *testing.T) {
				userRole, err := store.AddRoleToUser(ctx, user.ID, role.Name)
				require.Nil(t, err)
				require.Equal(t, user.ID, userRole.UserID)
				require.Equal(t, role.ID, userRole.RoleID)
			},
		},
		{
			"should_not_add_existing_role",
			func(t *testing.T) {
				_, err := store.AddRoleToUser(ctx, user.ID, role.Name)
				require.NotNil(t, err)
				require.ErrorIs(t, err, errs.ErrUserHasRole)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
