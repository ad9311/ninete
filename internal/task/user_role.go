package task

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ad9311/go-api-base/internal/errs"
)

// addRolesToUsersTask reads a CSV file from stdin and adds roles to users as specified in the file.
func (t *task) addRolesToUsersTask() error {
	ctx := context.Background()

	fmt.Print("Enter the CSV file path: ")
	reader := bufio.NewReader(os.Stdin)
	path, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	path = strings.TrimSpace(path)

	rows, err := openCSVFile(path)
	if err != nil {
		return err
	}

	rowsLength := len(rows)
	if rowsLength <= 1 {
		return ErrEmptyCSVFile
	}

	for _, row := range rows[1:] {
		err := t.addRoleToUser(ctx, row)
		if err != nil {
			return err
		}
	}

	return nil
}

// addRoleToUser adds a role to a user based on a row from the CSV file. Returns an error if the operation fails.
func (t *task) addRoleToUser(ctx context.Context, row []string) error {
	nCols := len(row)
	if nCols != 2 {
		return ErrWrongNumOfColumns
	}

	val, err := strconv.ParseInt(strings.TrimSpace(row[0]), 10, 32)
	if err != nil {
		return err
	}
	userID := int32(val)

	roleName := strings.TrimSpace(row[1])
	_, err = t.serviceStore.AddRoleToUser(ctx, int32(userID), roleName)
	if err != nil {
		if errors.Is(err, errs.ErrUserHasRole) {
			t.logger.Log("user %d already has the %s role\n", userID, roleName)

			return nil
		}

		t.logger.Log("could not add role %s to user with id %d\n", roleName, userID)

		return err
	}

	t.logger.Log("added role %s to user with id %d\n", roleName, userID)

	return nil
}
