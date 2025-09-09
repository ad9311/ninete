package service

import (
	"context"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/repo"
)

// CreateNewRole creates a new role with the given name, ensuring uniqueness.
// Returns an error if the role already exists or creation fails.
func (s *Store) CreateNewRole(ctx context.Context, name string) (repo.Role, error) {
	var role repo.Role

	if s.config.IsSafeEnv() {
		return role, errs.ErrServiceFuncNotAvailable
	}

	role, err := s.queries.InsertRole(ctx, name)
	if err != nil {
		return role, errs.HandlePgError(err)
	}

	return role, nil
}

// FindRoleByName retrieves a role from the database by its name.
// Returns an error if the role is not found or lookup fails.
func (s *Store) FindRoleByName(ctx context.Context, name string) (repo.Role, error) {
	role, err := s.queries.SelectRoleWhereName(ctx, name)
	if err != nil {
		return role, errs.HandlePgError(err)
	}

	return role, nil
}
