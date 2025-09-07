package task

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/repo"
)

// createAdminRoleTask ensures the 'admin' role exists in the database, creating it if necessary.
// Logs the result and returns any error encountered during creation.
func (t *task) createAdminRoleTask() error {
	ctx := context.Background()

	_, err := t.createNewRole(ctx, "admin")

	return err
}

// createNewRoleTask prompts the user for an admin role name via terminal and creates the role in the database.
func (t *task) createNewRoleTask() error {
	ctx := context.Background()

	fmt.Print("Enter admin role name: ")
	reader := bufio.NewReader(t.reader)
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	roleName := strings.TrimSpace(input)
	if roleName == "" {
		return ErrEmptyRoleName
	}
	_, err = t.createNewRole(ctx, roleName)

	return err
}

// createNewRole attempts to create a new role with the given name.
// If the role already exists, logs a message and returns the existing role.
// Returns any unique constraint violation error or other error encountered.
func (t *task) createNewRole(ctx context.Context, roleName string) (repo.Role, error) {
	role, err := t.serviceStore.CreateNewRole(ctx, roleName)
	if err != nil {
		if errors.Is(err, errs.ErrUniqueConstraintViolation) {
			t.logger.Log("%s role already exists!", roleName)

			return role, nil
		}

		return role, err
	}
	t.logger.Log("created %s role successfully!", roleName)

	return role, nil
}
