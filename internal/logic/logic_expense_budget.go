package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

// ExpenseBudgetParams is one category's budget as submitted by the budgets
// form. An Amount of zero means "no budget" and deletes any stored row.
type ExpenseBudgetParams struct {
	CategoryID int    `validate:"required"`
	Amount     uint64 `validate:"gte=0"`
}

func (s *Store) FindExpenseBudgets(ctx context.Context, userID int) ([]repo.ExpenseBudget, error) {
	return s.queries.SelectExpenseBudgetsByUser(ctx, userID)
}

func (s *Store) FindExpensesCategoryMonthTotals(
	ctx context.Context,
	filters repo.Filters,
) ([]repo.ExpenseCategoryMonthTotal, error) {
	return s.queries.SelectExpensesCategoryMonthTotals(ctx, filters)
}

// SaveExpenseBudgets writes every submitted category in one transaction: a
// non-zero amount upserts, a zero amount deletes. The form always posts every
// category, so a field the user cleared arrives here as zero and removes the
// budget rather than leaving a stale one behind.
func (s *Store) SaveExpenseBudgets(ctx context.Context, userID int, amountByCategoryID map[int]uint64) error {
	params := make([]ExpenseBudgetParams, 0, len(amountByCategoryID))

	for categoryID, amount := range amountByCategoryID {
		p := ExpenseBudgetParams{CategoryID: categoryID, Amount: amount}
		if err := s.ValidateStruct(p); err != nil {
			return err
		}

		params = append(params, p)
	}

	return s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		for _, p := range params {
			if p.Amount == 0 {
				if err := tq.DeleteExpenseBudget(ctx, userID, p.CategoryID); err != nil {
					return err
				}

				continue
			}

			if _, err := tq.UpsertExpenseBudget(ctx, repo.UpsertExpenseBudgetParams{
				UserID:     userID,
				CategoryID: p.CategoryID,
				Amount:     p.Amount,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Store) DeleteAllExpenseBudgets(ctx context.Context, userID int) error {
	return s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		return tq.DeleteAllExpenseBudgetsByUser(ctx, userID)
	})
}
