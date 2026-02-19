package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

type RecurrentExpenseParams struct {
	ExpenseBaseParams
	Period uint `validate:"required,gt=0"`
}

func (s *Store) FindRecurrentExpenses(
	ctx context.Context,
	opts repo.QueryOptions,
) ([]repo.RecurrentExpense, error) {
	recurrentExpenses, err := s.queries.SelectRecurrentExpenses(ctx, opts)
	if err != nil {
		return recurrentExpenses, err
	}

	return recurrentExpenses, nil
}

func (s *Store) CountRecurrentExpenses(ctx context.Context, filters repo.Filters) (int, error) {
	count, err := s.queries.CountRecurrentExpenses(ctx, filters)
	if err != nil {
		return count, err
	}

	return count, nil
}

func (s *Store) FindRecurrentExpense(ctx context.Context, id, userID int) (repo.RecurrentExpense, error) {
	recurrentExpense, err := s.queries.SelectRecurrentExpense(ctx, id, userID)
	if err != nil {
		return recurrentExpense, err
	}

	return recurrentExpense, nil
}

func (s *Store) CreateRecurrentExpense(
	ctx context.Context,
	userID int,
	params RecurrentExpenseParams,
) (repo.RecurrentExpense, error) {
	var recurrentExpense repo.RecurrentExpense

	if err := s.ValidateStruct(params); err != nil {
		return recurrentExpense, err
	}

	recurrentExpense, err := s.queries.InsertRecurrentExpense(ctx, repo.InsertRecurrentExpenseParams{
		UserID:      userID,
		CategoryID:  params.CategoryID,
		Description: params.Description,
		Amount:      params.Amount,
		Period:      params.Period,
	})
	if err != nil {
		return recurrentExpense, err
	}

	return recurrentExpense, nil
}

func (s *Store) UpdateRecurrentExpense(
	ctx context.Context,
	id, userID int,
	params RecurrentExpenseParams,
) (repo.RecurrentExpense, error) {
	var recurrentExpense repo.RecurrentExpense

	if err := s.ValidateStruct(params); err != nil {
		return recurrentExpense, err
	}

	recurrentExpense, err := s.queries.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
		ID:          id,
		UserID:      userID,
		CategoryID:  params.CategoryID,
		Description: params.Description,
		Amount:      params.Amount,
		Period:      params.Period,
	})
	if err != nil {
		return recurrentExpense, err
	}

	return recurrentExpense, nil
}

func (s *Store) DeleteRecurrentExpense(ctx context.Context, id, userID int) (int, error) {
	i, err := s.queries.DeleteRecurrentExpense(ctx, id, userID)
	if err != nil {
		return 0, err
	}

	return i, nil
}
