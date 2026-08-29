package logic_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestDeleteAllExpenses(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	user := s.CreateAuthUser(t, "acct_exp_user", "acct_exp_user@example.com", "password_1")
	otherUser := s.CreateAuthUser(t, "acct_exp_other", "acct_exp_other@example.com", "password_2")
	category := s.CreateCategory(t, "acct expense category")

	userExpense := s.CreateExpense(
		t, user.ID,
		newExpenseParams(category.ID, "acct user expense", 500, 1735689600, []string{"acct_tag_a"}),
	)
	otherExpense := s.CreateExpense(
		t, otherUser.ID,
		newExpenseParams(category.ID, "acct other expense", 600, 1735689600, []string{"acct_tag_b"}),
	)

	err := s.Store.DeleteAllExpenses(ctx, user.ID)
	require.NoError(t, err)

	userCount, err := s.Queries.CountExpensesByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Zero(t, userCount)

	// Taggings for the deleted expense must not be orphaned.
	orphaned, err := s.Queries.CountTaggingsByTarget(ctx, repo.TaggableTypeExpense, userExpense.ID)
	require.NoError(t, err)
	require.Zero(t, orphaned)

	// The user's tags themselves survive (only taggings are cleaned).
	tagCount, err := s.Queries.CountTagsByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 1, tagCount)

	// The other user's data is untouched.
	otherCount, err := s.Queries.CountExpensesByUser(ctx, otherUser.ID)
	require.NoError(t, err)
	require.Equal(t, 1, otherCount)

	otherTaggings, err := s.Queries.CountTaggingsByTarget(ctx, repo.TaggableTypeExpense, otherExpense.ID)
	require.NoError(t, err)
	require.Equal(t, 1, otherTaggings)
}

func TestDeleteAllRecurrentExpenses(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	user := s.CreateAuthUser(t, "acct_rec_user", "acct_rec_user@example.com", "password_1")
	otherUser := s.CreateAuthUser(t, "acct_rec_other", "acct_rec_other@example.com", "password_2")
	category := s.CreateCategory(t, "acct recurrent expense category")

	userParams := newRecurrentExpenseParams(category.ID, "acct user recurrent expense", 700, 1)
	userParams.Tags = []string{"acct_rec_tag_a"}
	userRecurrentExpense := s.CreateRecurrentExpense(t, user.ID, userParams)

	otherParams := newRecurrentExpenseParams(category.ID, "acct other recurrent expense", 800, 1)
	otherParams.Tags = []string{"acct_rec_tag_b"}
	otherRecurrentExpense := s.CreateRecurrentExpense(t, otherUser.ID, otherParams)

	err := s.Store.DeleteAllRecurrentExpenses(ctx, user.ID)
	require.NoError(t, err)

	userCount, err := s.Queries.CountRecurrentExpensesByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Zero(t, userCount)

	// Taggings key on the id rather than a foreign key, so a leftover row would
	// hand its tags to whichever recurrent expense SQLite gives that rowid next.
	orphaned, err := s.Queries.CountTaggingsByTarget(
		ctx,
		repo.TaggableTypeRecurrentExpense,
		userRecurrentExpense.ID,
	)
	require.NoError(t, err)
	require.Zero(t, orphaned)

	// The user's tags themselves survive (only taggings are cleaned).
	tagCount, err := s.Queries.CountTagsByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 1, tagCount)

	// The other user's data is untouched.
	otherCount, err := s.Queries.CountRecurrentExpensesByUser(ctx, otherUser.ID)
	require.NoError(t, err)
	require.Equal(t, 1, otherCount)

	otherTaggings, err := s.Queries.CountTaggingsByTarget(
		ctx,
		repo.TaggableTypeRecurrentExpense,
		otherRecurrentExpense.ID,
	)
	require.NoError(t, err)
	require.Equal(t, 1, otherTaggings)
}

func TestDeleteAllTagsCascadesTaggings(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	user := s.CreateAuthUser(t, "acct_tag_user", "acct_tag_user@example.com", "password_1")
	category := s.CreateCategory(t, "acct tag category")

	expense := s.CreateExpense(
		t, user.ID,
		newExpenseParams(category.ID, "acct tag expense", 500, 1735689600, []string{"acct_del_tag"}),
	)

	err := s.Store.DeleteAllTags(ctx, user.ID)
	require.NoError(t, err)

	tagCount, err := s.Queries.CountTagsByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Zero(t, tagCount)

	// Deleting tags cascades their taggings via the tag_id FK.
	taggings, err := s.Queries.CountTaggingsByTarget(ctx, repo.TaggableTypeExpense, expense.ID)
	require.NoError(t, err)
	require.Zero(t, taggings)
}

// Phase 0B of docs/spa-migration.md removed the macro, food and mood features,
// but deliberately left DeleteAllUserData clearing their tables: the tables live
// until Phase 8, and a user who tracked any of it before the removal still has
// rows there. Nothing else covers that now that the feature tests are gone, so
// this test seeds those tables through the repo layer directly — the logic-layer
// stores and the spec factories for them no longer exist. Deleting this test, or
// the four calls it guards, silently leaves personal data behind.
func TestDeleteAllUserData(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	user := s.CreateAuthUser(t, "acct_all_user", "acct_all_user@example.com", "password_1")
	otherUser := s.CreateAuthUser(t, "acct_all_other", "acct_all_other@example.com", "password_2")
	category := s.CreateCategory(t, "acct all category")

	seedDropped := func(userID int, suffix string) {
		err := s.Queries.WithTx(ctx, func(tq *repo.TxQueries) error {
			if _, err := tq.InsertMacroEntry(ctx, repo.InsertMacroEntryParams{
				UserID: userID, Name: "macro " + suffix, Kcal: 100, ProteinG: 10,
				CarbsG: 10, FatG: 5, Date: 1735689600, MealType: "lunch",
			}); err != nil {
				return err
			}
			if _, err := tq.UpsertMacroGoal(ctx, repo.UpsertMacroGoalParams{
				UserID: userID, Kcal: 2000, ProteinG: 150, CarbsG: 200, FatG: 70,
			}); err != nil {
				return err
			}
			if _, err := tq.InsertFood(ctx, repo.InsertFoodParams{
				UserID: userID, Name: "food " + suffix, Kcal: 100, ProteinG: 10,
				CarbsG: 10, FatG: 5,
			}); err != nil {
				return err
			}
			_, err := tq.InsertMoodEntry(ctx, repo.InsertMoodEntryParams{
				UserID: userID, Mood: "Happy", Notes: "notes " + suffix, LoggedAt: 1735689600,
			})

			return err
		})
		require.NoError(t, err)
	}

	seed := func(userID int, suffix string) {
		s.CreateExpense(t, userID, newExpenseParams(category.ID, "exp "+suffix, 500, 1735689600, []string{"tag_" + suffix}))
		s.CreateRecurrentExpense(t, userID, newRecurrentExpenseParams(category.ID, "rec "+suffix, 500, 1))
		seedDropped(userID, suffix)
	}

	seed(user.ID, "user")
	seed(otherUser.ID, "other")

	err := s.Store.DeleteAllUserData(ctx, user.ID)
	require.NoError(t, err)

	counts, err := s.Store.FindAccountDataCounts(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 0, counts.Expenses)
	require.Equal(t, 0, counts.RecurrentExpenses)
	require.Equal(t, 0, counts.Tags)

	requireDroppedCounts(t, s, user.ID, 0)

	otherCounts, err := s.Store.FindAccountDataCounts(ctx, otherUser.ID)
	require.NoError(t, err)
	require.Equal(t, 1, otherCounts.Expenses)
	require.Equal(t, 1, otherCounts.RecurrentExpenses)
	require.Equal(t, 1, otherCounts.Tags)

	requireDroppedCounts(t, s, otherUser.ID, 1)
}

// requireDroppedCounts asserts the row counts of the four tables Phase 8 will
// drop. They have no AccountDataCounts fields any more, so they are read from
// the repo layer.
func requireDroppedCounts(t *testing.T, s spec.Spec, userID, want int) {
	t.Helper()

	ctx := t.Context()

	macroEntries, err := s.Queries.CountMacroEntriesByUser(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, want, macroEntries)

	macroGoals, err := s.Queries.CountMacroGoalsByUser(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, want, macroGoals)

	foods, err := s.Queries.CountFoodsByUser(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, want, foods)

	moodEntries, err := s.Queries.CountMoodEntriesByUser(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, want, moodEntries)
}
