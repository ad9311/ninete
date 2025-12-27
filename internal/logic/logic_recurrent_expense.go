package logic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

type RecurrentExpenseParams struct {
	CategoryID  int    `json:"categoryId" validate:"required"`
	Description string `json:"description" validate:"required,min=3,max=50"`
	Amount      uint64 `json:"amount" validate:"required,gt=0"`
	Period      uint   `json:"period" validate:"required,gt=0,lt=25"`
}

type recurrentExpenseUpdateParams struct {
	Description string `validate:"required,min=3,max=50"`
	Amount      uint64 `validate:"required,gt=0"`
	Period      uint   `validate:"required,gt=0,lt=25"`
}

func (s *Store) FindRecurrentExpense(ctx context.Context, id, userID int) (repo.RecurrentExpense, error) {
	recurrentExpense, err := s.queries.SelectRecurrentExpense(ctx, id, userID)
	if err != nil {
		return recurrentExpense, HandleDBError(err)
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
		return recurrentExpense, HandleDBError(err)
	}

	return recurrentExpense, nil
}

func (s *Store) UpdateRecurrentExpense(
	ctx context.Context,
	params repo.UpdateRecurrentExpenseParams,
) (repo.RecurrentExpense, error) {
	var recurrentExpense repo.RecurrentExpense
	if err := s.ValidateStruct(recurrentExpenseUpdateParams{
		Description: params.Description,
		Amount:      params.Amount,
		Period:      params.Period,
	}); err != nil {
		return recurrentExpense, err
	}

	recurrentExpense, err := s.queries.UpdateRecurrentExpense(ctx, params)
	if err != nil {
		return recurrentExpense, HandleDBError(err)
	}

	return recurrentExpense, nil
}

func (s *Store) UpdateLastCopyCreatedAt(
	ctx context.Context,
	recurrent repo.RecurrentExpense,
	date int64,
) (repo.RecurrentExpense, error) {
	re, err := s.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
		ID:                recurrent.ID,
		UserID:            recurrent.UserID,
		Description:       recurrent.Description,
		Amount:            recurrent.Amount,
		Period:            recurrent.Period,
		LastCopyCreatedAt: sql.NullInt64{Int64: date, Valid: true},
	})
	if err != nil {
		return re, err
	}

	return re, nil
}

func (s *Store) CreateExpenseFromPeriod(
	ctx context.Context,
	recurrent repo.RecurrentExpense,
) (repo.Expense, error) {
	var expense repo.Expense

	if !recurrent.LastCopyCreatedAt.Valid {
		expense, err := s.createExpenseFromRecurrent(ctx, recurrent)
		if err != nil {
			return expense, err
		}

		return expense, nil
	}

	lastCopy := time.Unix(recurrent.LastCopyCreatedAt.Int64, 0).UTC()
	now := time.Now().UTC()

	timePassed := monthsBetween(lastCopy, now)
	if recurrent.Period > 24 {
		return expense, fmt.Errorf("%w, period is over 24 months", ErrValidationFailed)
	}

	if timePassed < int(recurrent.Period) {
		return expense, fmt.Errorf(
			"%w, cannot have two expenses from the same recurrent expense in the same period",
			ErrRecordAlreadyExist,
		)
	}

	expense, err := s.createExpenseFromRecurrent(ctx, recurrent)
	if err != nil {
		return expense, err
	}

	return expense, nil
}

func (s *Store) createExpenseFromRecurrent(
	ctx context.Context,
	recurrent repo.RecurrentExpense,
) (repo.Expense, error) {
	var expense repo.Expense

	nowUnix := time.Now().Unix()
	nowStr := prog.UnixToStringDate(nowUnix)
	expenseParams := ExpenseParams{
		CategoryID:  recurrent.CategoryID,
		Description: recurrent.Description,
		Amount:      recurrent.Amount,
		Date:        nowStr,
	}
	if err := s.ValidateStruct(expenseParams); err != nil {
		return expense, err
	}

	date, err := prog.StringToUnixDate(expenseParams.Date)
	if err != nil {
		return expense, err
	}

	err = s.queries.WithTx(ctx, func(txq *repo.TxQueries) error {
		var txErr error

		expense, txErr = txq.InsertExpense(ctx, repo.InsertExpenseParams{
			UserID:      recurrent.UserID,
			CategoryID:  recurrent.CategoryID,
			Description: recurrent.Description,
			Amount:      recurrent.Amount,
			Date:        date,
		})
		if txErr != nil {
			return HandleDBError(txErr)
		}

		_, txErr = txq.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
			ID:                recurrent.ID,
			UserID:            recurrent.UserID,
			Description:       recurrent.Description,
			Amount:            recurrent.Amount,
			Period:            recurrent.Period,
			LastCopyCreatedAt: sql.NullInt64{Int64: nowUnix, Valid: true},
		})
		if txErr != nil {
			return HandleDBError(txErr)
		}

		return nil
	})
	if err != nil {
		return expense, err
	}

	return expense, nil
}

func monthsBetween(a, b time.Time) int {
	if a.After(b) {
		a, b = b, a
	}

	months := (b.Year()-a.Year())*12 + int(b.Month()-a.Month())

	if b.Day() < a.Day() {
		months--
	}

	if months < 0 {
		return 0
	}

	return months
}
