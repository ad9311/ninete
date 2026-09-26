package logic

import (
	"context"
	"strings"

	"github.com/ad9311/ninete/internal/repo"
)

type ExpenseParams struct {
	ExpenseBaseParams
	Date int64 `validate:"required,gt=0"`
	// Note is optional free text, shown only on the expense's own page. max
	// counts characters (runes), not bytes. Trimmed before validation, see
	// normalizeExpenseParams.
	Note string   `validate:"max=255"`
	Tags []string `validate:"-"`
}

// normalizeExpenseParams trims the note so surrounding whitespace neither
// counts toward the limit nor survives as a note that looks empty.
func normalizeExpenseParams(params ExpenseParams) ExpenseParams {
	params.Note = strings.TrimSpace(params.Note)

	return params
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
	tags, err := s.queries.SelectTagsForTaggable(ctx, repo.TaggableExpense(), expenseID, userID)
	if err != nil {
		return tags, err
	}

	return tags, nil
}

func (s *Store) CreateExpense(ctx context.Context, userID int, params ExpenseParams) (repo.Expense, error) {
	var expense repo.Expense

	params = normalizeExpenseParams(params)

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
			Note:        params.Note,
		})
		if txErr != nil {
			return txErr
		}

		return s.replaceTagsTx(ctx, tq, repo.TaggableExpense(), expense.ID, userID, params.Tags)
	})
	if err != nil {
		return expense, err
	}

	return expense, nil
}

func (s *Store) UpdateExpense(ctx context.Context, id, userID int, params ExpenseParams) (repo.Expense, error) {
	var expense repo.Expense

	params = normalizeExpenseParams(params)

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
			Note:        params.Note,
		})
		if txErr != nil {
			return txErr
		}

		return s.replaceTagsTx(ctx, tq, repo.TaggableExpense(), expense.ID, userID, params.Tags)
	})
	if err != nil {
		return expense, err
	}

	return expense, nil
}

// DeleteExpense removes the record and its taggings together. The taggings row
// points at the id, not at a foreign key, so leaving it behind would hand its
// tags to whichever expense SQLite gives that rowid next.
func (s *Store) DeleteExpense(ctx context.Context, id, userID int) (int, error) {
	var deletedID int

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		deletedID, txErr = tq.DeleteExpense(ctx, id, userID)
		if txErr != nil {
			return txErr
		}

		return tq.DeleteTaggingsByTarget(ctx, repo.TaggableExpense(), deletedID)
	})
	if err != nil {
		return 0, err
	}

	return deletedID, nil
}

func (s *Store) DeleteAllExpenses(ctx context.Context, userID int) error {
	return s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		return tq.DeleteAllExpensesByUser(ctx, userID)
	})
}

func (s *Store) FindExpensesCategoryTotals(
	ctx context.Context,
	filters repo.Filters,
) ([]repo.ExpenseCategoryTotal, error) {
	return s.queries.SelectExpensesCategoryTotals(ctx, filters)
}
