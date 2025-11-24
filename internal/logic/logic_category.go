package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

func (s *Store) CreateCategory(ctx context.Context, name, uid string) (repo.Category, error) {
	category, err := s.queries.InserCategory(ctx, name, uid)
	if err != nil {
		return category, HandleDBError(err)
	}

	return category, nil
}
