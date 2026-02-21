package logic

import (
	"context"
	"database/sql"
	"time"

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

func (s *Store) CopyDueRecurrentExpenses(ctx context.Context, now time.Time) (int, error) {
	nowUnix := now.Unix()
	expenseDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Unix()

	recurrentExpenses, err := s.queries.SelectAllDueRecurrentExpenses(ctx, nowUnix)
	if err != nil {
		return 0, err
	}

	if len(recurrentExpenses) == 0 {
		return 0, nil
	}

	copied := 0
	for _, re := range recurrentExpenses {
		if err := s.copyRecurrentExpense(ctx, re, expenseDate); err != nil {
			s.app.Logger.Errorf("failed to copy recurrent expense [id=%d]: %v", re.ID, err)

			continue
		}

		copied++
	}

	return copied, nil
}

func (s *Store) copyRecurrentExpense(ctx context.Context, re repo.RecurrentExpense, expenseDate int64) error {
	return s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		_, err := tq.InsertExpense(ctx, repo.InsertExpenseParams{
			UserID:      re.UserID,
			CategoryID:  re.CategoryID,
			Description: re.Description,
			Amount:      re.Amount,
			Date:        expenseDate,
		})
		if err != nil {
			return err
		}

		_, err = tq.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
			ID:                re.ID,
			UserID:            re.UserID,
			CategoryID:        re.CategoryID,
			Description:       re.Description,
			Amount:            re.Amount,
			Period:            re.Period,
			LastCopyCreatedAt: sql.NullInt64{Int64: expenseDate, Valid: true},
		})

		return err
	})
}
