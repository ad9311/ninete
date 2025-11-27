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

func (s *Store) FindCategories(ctx context.Context) ([]repo.Category, error) {
	categories, err := s.queries.SelectCategories(ctx)
	if err != nil {
		return categories, HandleDBError(err)
	}

	return categories, nil
}
