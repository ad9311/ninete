package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

func (s *Store) CreateCategory(ctx context.Context, name string) (repo.Category, error) {
	var category repo.Category

	uid := prog.ToLowerCamel(name)
	category, err := s.queries.InserCategory(ctx, name, uid)
	if err != nil {
		return category, err
	}

	return category, nil
}

func (s *Store) FindCategories(ctx context.Context) ([]repo.Category, error) {
	categories, err := s.queries.SelectCategories(ctx)
	if err != nil {
		return categories, err
	}

	return categories, nil
}
