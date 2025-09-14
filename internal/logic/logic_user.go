package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ad9311/ninete/internal/repo"
)

func (s *Store) FindUserByEmail(ctx context.Context, email string) (repo.User, error) {
	var user repo.User

	user, err := s.queries.SelectUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, ErrNotFound
		}

		return user, err
	}

	return user, nil
}
