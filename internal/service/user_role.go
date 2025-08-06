package service

import (
	"context"
	"slices"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/repo"
)

// AddRoleToUser links a user to a role by creating a user_role record. Returns an error if the user already has the role or if any step fails.
func (s *Store) AddRoleToUser(
	ctx context.Context,
	userID int32,
	roleName string,
) (repo.UserRole, error) {
	var userRole repo.UserRole

	if s.config.IsSafeEnv() {
		return userRole, errs.ErrServiceFuncNotAvailable
	}

	role, err := s.FindRoleByName(ctx, roleName)
	if err != nil {
		return userRole, err
	}

	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return userRole, err
	}

	if slices.Contains(roles, role) {
		return userRole, errs.ErrUserHasRole
	}

	userRole, err = s.queries.InsertUserRole(ctx, repo.InsertUserRoleParams{
		UserID: userID,
		RoleID: role.ID,
	})
	if err != nil {
		return userRole, errs.HandlePgError(err)
	}

	return userRole, nil
}
