package service

import (
	"context"
	"slices"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/repo"
)

// FindUserByID retrieves a user by their ID and returns a SafeUser struct.
// Returns an error if the user is not found or lookup fails.
func (s *Store) FindUserByID(ctx context.Context, userID int32) (SafeUser, error) {
	var safeUser SafeUser

	user, err := s.queries.SelectUserWhereId(ctx, userID)
	if err != nil {
		return safeUser, errs.HandlePgError(err)
	}

	safeUser.ID = user.ID
	safeUser.Username = user.Username
	safeUser.Email = user.Email

	return safeUser, nil
}

// FindUserByEmail retrieves a user by their email address. Returns an error if the user is not found or lookup fails.
func (s *Store) FindUserByEmail(ctx context.Context, email string) (repo.User, error) {
	user, err := s.queries.SelectUserWhereEmail(ctx, email)
	if err != nil {
		return user, errs.HandlePgError(err)
	}

	return user, nil
}

// GetUserRoles returns all roles associated with the specified user ID. Returns an error if the lookup fails.
func (s *Store) GetUserRoles(ctx context.Context, userID int32) ([]repo.Role, error) {
	roles, err := s.queries.SelectRolesWhereUserID(ctx, userID)
	if err != nil {
		return roles, errs.HandlePgError(err)
	}

	return roles, nil
}

// UserRoleNames returns a sorted slice of role names extracted from the given slice of repo.Role.
func UserRoleNames(roles []repo.Role) []string {
	var names []string

	for _, r := range roles {
		names = append(names, r.Name)
	}

	slices.Sort(names)

	return names
}
