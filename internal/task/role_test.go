package task

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCreateAdminRoleTask(t *testing.T) {
	tbuff := newTestBuffer()
	task := newTaskFactory(t, &tbuff.stdOut, &tbuff.stdOut)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_create_admin_role",
			func(t *testing.T) {
				err := task.createAdminRoleTask()
				require.Nil(t, err)

				checkRole(t, task, "admin")
			},
		},
		{
			"should_not_duplicate_admin_role",
			func(t *testing.T) {
				err := task.createAdminRoleTask()
				require.Nil(t, err)

				want := "admin role already exists!"
				require.Contains(t, tbuff.stdOut.String(), want)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
		tbuff.clear()
	}
}

func TestCreateNewRoleTask(t *testing.T) {
	tbuff := newTestBuffer()
	task := newTaskFactory(t, &tbuff.stdOut, &tbuff.stdErr)

	uniqueRoleName := service.FactoryUsername()

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_create_a_new_role",
			func(t *testing.T) {
				task.reader = strings.NewReader(uniqueRoleName + "\n")
				err := task.createNewRoleTask()

				require.Nil(t, err)

				checkRole(t, task, uniqueRoleName)
			},
		},
		{
			"should_trim_surrounding_spaces",
			func(t *testing.T) {
				roleName := service.FactoryUsername()
				task.reader = strings.NewReader(" " + roleName + "     " + "\n")
				err := task.createNewRoleTask()

				require.Nil(t, err)

				checkRole(t, task, roleName)
			},
		},
		{
			"should_output_already_created",
			func(t *testing.T) {
				task.reader = strings.NewReader(uniqueRoleName + "\n")
				err := task.createNewRoleTask()
				require.Nil(t, err)

				want := fmt.Sprintf("%s role already exists!", uniqueRoleName)
				require.Contains(t, tbuff.stdOut.String(), want)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
		tbuff.clear()
	}
}

func checkRole(t *testing.T, task *task, roleName string) {
	ctx := context.Background()

	role, err := task.serviceStore.FindRoleByName(ctx, roleName)
	require.Nil(t, err)
	require.Positive(t, role.ID)
	require.Equal(t, roleName, role.Name)
}
