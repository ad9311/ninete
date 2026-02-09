package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

func (s *Store) FindUserByEmail(ctx context.Context, email string) (repo.User, error) {
	var user repo.User

	user, err := s.queries.SelectUserByEmail(ctx, email)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *Store) FindUser(ctx context.Context, id int) (repo.User, error) {
	var user repo.User

	user, err := s.queries.SelectUser(ctx, id)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *Store) CreateUser(ctx context.Context, params repo.InsertUserParams) (repo.User, error) {
	var user repo.User

	user, err := s.queries.InsertUser(ctx, params)
	if err != nil {
		return user, err
	}

	return user, nil
}
