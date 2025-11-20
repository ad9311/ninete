package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

func (s *Store) CreateExpense(ctx context.Context, params repo.InsertExpenseParams) (repo.Expense, error) {
	var expense repo.Expense

	if err := s.ValidateStruct(params); err != nil {
		return expense, err
	}

	expense, err := s.queries.InsertExpense(ctx, params)
	if err != nil {
		return expense, HandleDBError(err)
	}

	return expense, nil
}

func (s *Store) UpdateExpense(ctx context.Context, params repo.UpdateExpenseParams) (repo.Expense, error) {
	var expense repo.Expense

	if err := s.ValidateStruct(params); err != nil {
		return expense, err
	}

	expense, err := s.queries.UpdateExpense(ctx, params)
	if err != nil {
		return expense, HandleDBError(err)
	}

	return expense, nil
}

func (s *Store) DeleteExpense(ctx context.Context, id int) (repo.Expense, error) {
	var expense repo.Expense

	expense, err := s.queries.DeleteExpense(ctx, id)
	if err != nil {
		return expense, HandleDBError(err)
	}

	return expense, nil
}
