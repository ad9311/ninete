package logic

import (
	"context"
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
	nowStr := prog.UnixToStringDate(time.Now().Unix())

	expense, err := s.CreateExpense(ctx, recurrent.UserID, ExpenseParams{
		CategoryID:  recurrent.CategoryID,
		Description: recurrent.Description,
		Amount:      recurrent.Amount,
		Date:        nowStr,
	})
	if err != nil {
		return expense, HandleDBError(err)
	}

	_, err = s.queries.UpdateLastCopyCreated(ctx, recurrent.ID, time.Now().Unix())
	if err != nil {
		return expense, HandleDBError(err)
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
