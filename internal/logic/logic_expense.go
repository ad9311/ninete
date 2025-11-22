package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

type ExpenseParams struct {
	CategoryID  int    `json:"categoryId" validate:"required"`
	Description string `json:"description" validate:"required,min=3,max=50"`
	Amount      uint64 `json:"amount" validate:"required,gt=0"`
	Date        int64  `json:"date" validate:"required"`
}

func (s *Store) FindExpenses(ctx context.Context, opts repo.QueryOptions) ([]repo.Expense, error) {
	expenses, err := s.queries.SelectExpenses(ctx, opts)
	if err != nil {
		return expenses, HandleDBError(err)
	}

	return expenses, nil
}

func (s *Store) CountExpenses(ctx context.Context, filters repo.Filters) (int, error) {
	count, err := s.queries.CountExpenses(ctx, filters)
	if err != nil {
		return count, HandleDBError(err)
	}

	return count, nil
}

func (s *Store) FindExpense(ctx context.Context, id, userID int) (repo.Expense, error) {
	expense, err := s.queries.SelectExpense(ctx, id, userID)
	if err != nil {
		return expense, HandleDBError(err)
	}

	return expense, nil
}

func (s *Store) CreateExpense(ctx context.Context, userID int, params ExpenseParams) (repo.Expense, error) {
	var expense repo.Expense

	if err := s.ValidateStruct(params); err != nil {
		return expense, err
	}

	expense, err := s.queries.InsertExpense(ctx, repo.InsertExpenseParams{
		UserID:      userID,
		CategoryID:  params.CategoryID,
		Description: params.Description,
		Amount:      params.Amount,
		Date:        params.Date,
	})
	if err != nil {
		return expense, HandleDBError(err)
	}

	return expense, nil
}

func (s *Store) UpdateExpense(ctx context.Context, id int, params ExpenseParams) (repo.Expense, error) {
	var expense repo.Expense

	if err := s.ValidateStruct(params); err != nil {
		return expense, err
	}

	expense, err := s.queries.UpdateExpense(ctx, repo.UpdateExpenseParams{
		ID:          id,
		CategoryID:  params.CategoryID,
		Description: params.Description,
		Amount:      params.Amount,
		Date:        params.Date,
	})
	if err != nil {
		return expense, HandleDBError(err)
	}

	return expense, nil
}

func (s *Store) DeleteExpense(ctx context.Context, id int) (int, error) {
	i, err := s.queries.DeleteExpense(ctx, id)
	if err != nil {
		return 0, HandleDBError(err)
	}

	return i, nil
}
