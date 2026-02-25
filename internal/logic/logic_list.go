package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

type ListParams struct {
	Name string `validate:"required,max=100"`
}

func (s *Store) FindLists(ctx context.Context, opts repo.QueryOptions) ([]repo.List, error) {
	lists, err := s.queries.SelectLists(ctx, opts)
	if err != nil {
		return lists, err
	}

	return lists, nil
}

func (s *Store) CountLists(ctx context.Context, filters repo.Filters) (int, error) {
	count, err := s.queries.CountLists(ctx, filters)
	if err != nil {
		return count, err
	}

	return count, nil
}

func (s *Store) FindList(ctx context.Context, id, userID int) (repo.List, error) {
	list, err := s.queries.SelectList(ctx, id, userID)
	if err != nil {
		return list, err
	}

	return list, nil
}

func (s *Store) CreateList(ctx context.Context, userID int, params ListParams) (repo.List, error) {
	var list repo.List

	if err := s.ValidateStruct(params); err != nil {
		return list, err
	}

	list, err := s.queries.InsertList(ctx, repo.InsertListParams{
		UserID: userID,
		Name:   params.Name,
	})
	if err != nil {
		return list, err
	}

	return list, nil
}

func (s *Store) UpdateList(ctx context.Context, id, userID int, params ListParams) (repo.List, error) {
	var list repo.List

	if err := s.ValidateStruct(params); err != nil {
		return list, err
	}

	list, err := s.queries.UpdateList(ctx, userID, repo.UpdateListParams{
		ID:   id,
		Name: params.Name,
	})
	if err != nil {
		return list, err
	}

	return list, nil
}

func (s *Store) DeleteList(ctx context.Context, id, userID int) (int, error) {
	i, err := s.queries.DeleteList(ctx, id, userID)
	if err != nil {
		return 0, err
	}

	return i, nil
}
