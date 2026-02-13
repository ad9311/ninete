package logic

import (
	"context"
	"strings"

	"github.com/ad9311/ninete/internal/repo"
)

type TagParams struct {
	Name string `validate:"required,min=1,max=50"`
}

func (s *Store) FindTags(ctx context.Context, opts repo.QueryOptions) ([]repo.Tag, error) {
	tags, err := s.queries.SelectTags(ctx, opts)
	if err != nil {
		return tags, err
	}

	return tags, nil
}

func (s *Store) CreateTag(ctx context.Context, userID int, params TagParams) (repo.Tag, error) {
	var tag repo.Tag

	params.Name = strings.TrimSpace(params.Name)
	if err := s.ValidateStruct(params); err != nil {
		return tag, err
	}

	tag, err := s.queries.InsertTag(ctx, repo.InsertTagParams{
		UserID: userID,
		Name:   params.Name,
	})
	if err != nil {
		return tag, err
	}

	return tag, nil
}

func (s *Store) DeleteTag(ctx context.Context, id, userID int) (int, error) {
	i, err := s.queries.DeleteTag(ctx, id, userID)
	if err != nil {
		return 0, err
	}

	return i, nil
}
