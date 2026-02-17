package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

type ExpenseParams struct {
	expenseBaseParams
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

func ExtractTagNames(tags []repo.Tag) []string {
	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	return tagNames
}

func (s *Store) FindExpenseTagRows(
	ctx context.Context,
	expenseIDs []int,
	userID int,
) ([]repo.ExpenseTagRow, error) {
	rows, err := s.queries.SelectExpenseTagRows(ctx, expenseIDs, userID)
	if err != nil {
		return rows, err
	}

	return rows, nil
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

		return s.replaceExpenseTagsTx(ctx, tq, expense.ID, userID, params.Tags)
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

		return s.replaceExpenseTagsTx(ctx, tq, expense.ID, userID, params.Tags)
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

func (s *Store) replaceExpenseTagsTx(
	ctx context.Context,
	tq *repo.TxQueries,
	expenseID int,
	userID int,
	tagNames []string,
) error {
	if err := tq.DeleteTaggingsByTarget(ctx, repo.TaggableTypeExpense, expenseID); err != nil {
		return err
	}

	if len(tagNames) == 0 {
		return nil
	}

	tags, err := s.ensureTagsForUserTx(ctx, tq, userID, tagNames)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		err := tq.InsertOrIgnoreTagging(ctx, repo.InsertTaggingParams{
			TagID:        tag.ID,
			TaggableID:   expenseID,
			TaggableType: repo.TaggableTypeExpense,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
