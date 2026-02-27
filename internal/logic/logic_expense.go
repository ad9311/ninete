package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

type ExpenseParams struct {
	ExpenseBaseParams
	Date int64    `validate:"required,gt=0"`
	Tags []string `validate:"-"`
}

func (s *Store) FindExpenses(ctx context.Context, opts repo.QueryOptions) ([]repo.Expense, error) {
	expenses, err := s.queries.SelectExpenses(ctx, opts)
	if err != nil {
		return expenses, err
	}

	return expenses, nil
}

func (s *Store) CountExpenses(ctx context.Context, filters repo.Filters) (int, error) {
	count, err := s.queries.CountExpenses(ctx, filters)
	if err != nil {
		return count, err
	}

	return count, nil
}

func (s *Store) FindExpense(ctx context.Context, id, userID int) (repo.Expense, error) {
	expense, err := s.queries.SelectExpense(ctx, id, userID)
	if err != nil {
		return expense, err
	}

	return expense, nil
}

func (s *Store) FindExpenseTags(ctx context.Context, expenseID, userID int) ([]repo.Tag, error) {
	tags, err := s.queries.SelectExpenseTags(ctx, expenseID, userID)
	if err != nil {
		return tags, err
	}

	return tags, nil
}

func (s *Store) CreateExpense(ctx context.Context, userID int, params ExpenseParams) (repo.Expense, error) {
	var expense repo.Expense

	if err := s.ValidateStruct(params); err != nil {
		return expense, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		expense, txErr = tq.InsertExpense(ctx, repo.InsertExpenseParams{
			UserID:      userID,
			CategoryID:  params.CategoryID,
			Description: params.Description,
			Amount:      params.Amount,
			Date:        params.Date,
		})
		if txErr != nil {
			return txErr
		}

		return s.replaceTagsTx(ctx, tq, repo.TaggableTypeExpense, expense.ID, userID, params.Tags)
	})
	if err != nil {
		return expense, err
	}

	return expense, nil
}

func (s *Store) UpdateExpense(ctx context.Context, id, userID int, params ExpenseParams) (repo.Expense, error) {
	var expense repo.Expense

	if err := s.ValidateStruct(params); err != nil {
		return expense, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		expense, txErr = tq.UpdateExpense(ctx, userID, repo.UpdateExpenseParams{
			ID:          id,
			CategoryID:  params.CategoryID,
			Description: params.Description,
			Amount:      params.Amount,
			Date:        params.Date,
		})
		if txErr != nil {
			return txErr
		}

		return s.replaceTagsTx(ctx, tq, repo.TaggableTypeExpense, expense.ID, userID, params.Tags)
	})
	if err != nil {
		return expense, err
	}

	return expense, nil
}

func (s *Store) DeleteExpense(ctx context.Context, id, userID int) (int, error) {
	i, err := s.queries.DeleteExpense(ctx, id, userID)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (s *Store) FindExpensesCategoryTotals(
	ctx context.Context,
	filters repo.Filters,
) ([]repo.ExpenseCategoryTotal, error) {
	return s.queries.SelectExpensesCategoryTotals(ctx, filters)
}
