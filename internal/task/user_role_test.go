package task

import (
	"context"
	"fmt"
	"testing"

	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAddRoleToUser(t *testing.T) {
	ctx := context.Background()
	tbuff := newTestBuffer()
	task := newTaskFactory(t, &tbuff.stdOut, &tbuff.stdErr)

	user := task.serviceStore.FactoryUser(ctx, t, service.RegistrationParams{})
	roleName := service.FactoryUsername()
	_, err := task.createNewRole(ctx, roleName)
	require.Nil(t, err)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_add_role_to_user",
			func(t *testing.T) {
				userIDstr := fmt.Sprintf("%d", user.ID)
				row := []string{userIDstr, roleName}
				err := task.addRoleToUser(ctx, row)
				require.Nil(t, err)

				roles, err := task.serviceStore.GetUserRoles(ctx, user.ID)
				require.Nil(t, err)
				roleNames := service.UserRoleNames(roles)
				require.Contains(t, roleNames, roleName)

				want := fmt.Sprintf("added role %s to user with id %d", roleName, user.ID)
				require.Contains(t, tbuff.stdOut.String(), want)
			},
		},
		{
			"should_not_add_role_when_user_has_it",
			func(t *testing.T) {
				userIDstr := fmt.Sprintf("%d", user.ID)
				row := []string{userIDstr, roleName}
				err := task.addRoleToUser(ctx, row)
				require.Nil(t, err)

				roles, err := task.serviceStore.GetUserRoles(ctx, user.ID)
				require.Nil(t, err)
				require.Len(t, roles, 1)

				want := fmt.Sprintf("user %d already has the %s role", user.ID, roleName)
				require.Contains(t, tbuff.stdOut.String(), want)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
		tbuff.clear()
	}
}
