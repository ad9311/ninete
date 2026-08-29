package logic

import (
	"context"
	"time"

	"github.com/ad9311/ninete/internal/repo"
)

type RecurrentExpenseParams struct {
	ExpenseBaseParams
	Period uint `validate:"required,gt=0"`
	// OccurrenceLimit caps how many expenses this recurrent expense generates
	// before it archives itself. Zero means unlimited.
	OccurrenceLimit uint     `validate:"-"`
	Tags            []string `validate:"-"`
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

func (s *Store) FindRecurrentExpenseTags(ctx context.Context, recurrentExpenseID, userID int) ([]repo.Tag, error) {
	tags, err := s.queries.SelectTagsForTaggable(
		ctx,
		repo.TaggableRecurrentExpense(),
		recurrentExpenseID,
		userID,
	)
	if err != nil {
		return tags, err
	}

	return tags, nil
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

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		recurrentExpense, txErr = tq.InsertRecurrentExpense(ctx, repo.InsertRecurrentExpenseParams{
			UserID:          userID,
			CategoryID:      params.CategoryID,
			Description:     params.Description,
			Amount:          params.Amount,
			Period:          params.Period,
			OccurrenceLimit: params.OccurrenceLimit,
		})
		if txErr != nil {
			return txErr
		}

		return s.replaceTagsTx(
			ctx,
			tq,
			repo.TaggableRecurrentExpense(),
			recurrentExpense.ID,
			userID,
			params.Tags,
		)
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

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		recurrentExpense, txErr = tq.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
			ID:              id,
			UserID:          userID,
			CategoryID:      params.CategoryID,
			Description:     params.Description,
			Amount:          params.Amount,
			Period:          params.Period,
			OccurrenceLimit: params.OccurrenceLimit,
		})
		if txErr != nil {
			return txErr
		}

		return s.replaceTagsTx(
			ctx,
			tq,
			repo.TaggableRecurrentExpense(),
			recurrentExpense.ID,
			userID,
			params.Tags,
		)
	})
	if err != nil {
		return recurrentExpense, err
	}

	return recurrentExpense, nil
}

// DeleteRecurrentExpense removes the record and its taggings together. The
// taggings row points at the id, not at a foreign key, so leaving it behind
// would hand its tags to whichever recurrent expense SQLite gives that rowid next.
func (s *Store) DeleteRecurrentExpense(ctx context.Context, id, userID int) (int, error) {
	var deletedID int

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		deletedID, txErr = tq.DeleteRecurrentExpense(ctx, id, userID)
		if txErr != nil {
			return txErr
		}

		return tq.DeleteTaggingsByTarget(ctx, repo.TaggableRecurrentExpense(), deletedID)
	})
	if err != nil {
		return 0, err
	}

	return deletedID, nil
}

func (s *Store) DeleteAllRecurrentExpenses(ctx context.Context, userID int) error {
	return s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		return tq.DeleteAllRecurrentExpensesByUser(ctx, userID)
	})
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
		expense, err := tq.InsertExpense(ctx, repo.InsertExpenseParams{
			UserID:      re.UserID,
			CategoryID:  re.CategoryID,
			Description: re.Description,
			Amount:      re.Amount,
			Date:        expenseDate,
		})
		if err != nil {
			return err
		}

		err = tq.CopyTaggings(
			ctx,
			repo.TaggableRecurrentExpense(),
			re.ID,
			repo.TaggableExpense(),
			expense.ID,
		)
		if err != nil {
			return err
		}

		_, err = tq.RecordRecurrentExpenseOccurrence(ctx, re.ID, re.UserID, expenseDate)

		return err
	})
}

// UnarchiveRecurrentExpense clears the archived flag and resets the occurrence
// counter, so the cron job starts a fresh run of "occurrence_limit" copies.
func (s *Store) UnarchiveRecurrentExpense(ctx context.Context, id, userID int) (repo.RecurrentExpense, error) {
	var recurrentExpense repo.RecurrentExpense

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		recurrentExpense, txErr = tq.UnarchiveRecurrentExpense(ctx, id, userID)

		return txErr
	})
	if err != nil {
		return recurrentExpense, err
	}

	return recurrentExpense, nil
}
