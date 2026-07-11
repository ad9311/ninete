package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

// AccountDataCounts holds the number of records a user owns per model type,
// used to populate the account "danger zone" page.
type AccountDataCounts struct {
	Expenses          int
	RecurrentExpenses int
	MacroEntries      int
	MacroGoals        int
	Foods             int
	MoodEntries       int
	Tags              int
}

// FindAccountDataCounts returns per-model record counts for the given user.
func (s *Store) FindAccountDataCounts(ctx context.Context, userID int) (AccountDataCounts, error) {
	var counts AccountDataCounts
	var err error

	if counts.Expenses, err = s.queries.CountExpensesByUser(ctx, userID); err != nil {
		return counts, err
	}
	if counts.RecurrentExpenses, err = s.queries.CountRecurrentExpensesByUser(ctx, userID); err != nil {
		return counts, err
	}
	if counts.MacroEntries, err = s.queries.CountMacroEntriesByUser(ctx, userID); err != nil {
		return counts, err
	}
	if counts.MacroGoals, err = s.queries.CountMacroGoalsByUser(ctx, userID); err != nil {
		return counts, err
	}
	if counts.Foods, err = s.queries.CountFoodsByUser(ctx, userID); err != nil {
		return counts, err
	}
	if counts.MoodEntries, err = s.queries.CountMoodEntriesByUser(ctx, userID); err != nil {
		return counts, err
	}
	if counts.Tags, err = s.queries.CountTagsByUser(ctx, userID); err != nil {
		return counts, err
	}

	return counts, nil
}

// DeleteAllUserData removes every record owned by the user across all model
// types in a single transaction (all-or-nothing). Tags are deleted last so any
// remaining taggings cascade away via their tag_id foreign key.
func (s *Store) DeleteAllUserData(ctx context.Context, userID int) error {
	return s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		if err := tq.DeleteAllExpensesByUser(ctx, userID); err != nil {
			return err
		}
		if err := tq.DeleteAllRecurrentExpensesByUser(ctx, userID); err != nil {
			return err
		}
		if err := tq.DeleteAllMacroEntriesByUser(ctx, userID); err != nil {
			return err
		}
		if err := tq.DeleteAllMacroGoalsByUser(ctx, userID); err != nil {
			return err
		}
		if err := tq.DeleteAllFoodsByUser(ctx, userID); err != nil {
			return err
		}
		if err := tq.DeleteAllMoodEntriesByUser(ctx, userID); err != nil {
			return err
		}

		return tq.DeleteAllTagsByUser(ctx, userID)
	})
}
