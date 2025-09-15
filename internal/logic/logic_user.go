package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

func (s *Store) FindUserByEmail(ctx context.Context, email string) (repo.User, error) {
	var user repo.User

	user, err := s.queries.SelectUserByEmail(ctx, email)
	if err != nil {
		return user, HandleDBError(err)
	}

	return user, nil
}

func (s *Store) FindUserByID(ctx context.Context, id int) (repo.User, error) {
	var user repo.User

	user, err := s.queries.SelectUserByID(ctx, id)
	if err != nil {
		return user, HandleDBError(err)
	}

	return user, nil
}
