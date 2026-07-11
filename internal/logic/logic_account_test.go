package logic_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/logic"
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

func TestDeleteAllMoodEntries(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	user := s.CreateAuthUser(t, "acct_mood_user", "acct_mood_user@example.com", "password_1")
	otherUser := s.CreateAuthUser(t, "acct_mood_other", "acct_mood_other@example.com", "password_2")

	userMood := s.CreateMoodEntry(
		t, user.ID,
		newMoodEntryParams("Happy", "acct notes", 1735689600, []string{"acct_mood_tag"}),
	)
	s.CreateMoodEntry(
		t, otherUser.ID,
		newMoodEntryParams("Calm", "acct other notes", 1735689600, []string{"acct_mood_tag_b"}),
	)

	err := s.Store.DeleteAllMoodEntries(ctx, user.ID)
	require.NoError(t, err)

	userCount, err := s.Queries.CountMoodEntriesByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Zero(t, userCount)

	orphaned, err := s.Queries.CountTaggingsByTarget(ctx, repo.TaggableTypeMoodEntry, userMood.ID)
	require.NoError(t, err)
	require.Zero(t, orphaned)

	otherCount, err := s.Queries.CountMoodEntriesByUser(ctx, otherUser.ID)
	require.NoError(t, err)
	require.Equal(t, 1, otherCount)
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

func TestDeleteAllUserData(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	user := s.CreateAuthUser(t, "acct_all_user", "acct_all_user@example.com", "password_1")
	otherUser := s.CreateAuthUser(t, "acct_all_other", "acct_all_other@example.com", "password_2")
	category := s.CreateCategory(t, "acct all category")

	seed := func(userID int, suffix string) {
		s.CreateExpense(t, userID, newExpenseParams(category.ID, "exp "+suffix, 500, 1735689600, []string{"tag_" + suffix}))
		s.CreateRecurrentExpense(t, userID, newRecurrentExpenseParams(category.ID, "rec "+suffix, 500, 1))
		s.CreateMacroEntry(t, userID, newMacroEntryParams("macro "+suffix, 100, 10, 10, 5, 1735689600))
		s.SaveMacroGoal(t, userID, logic.MacroGoalParams{Kcal: 2000, ProteinG: 150, CarbsG: 200, FatG: 70})
		s.CreateFood(t, userID, newFoodParams("food "+suffix, 100, 10, 10, 5))
		s.CreateMoodEntry(t, userID, newMoodEntryParams("Happy", "notes "+suffix, 1735689600, []string{"mtag_" + suffix}))
	}

	seed(user.ID, "user")
	seed(otherUser.ID, "other")

	err := s.Store.DeleteAllUserData(ctx, user.ID)
	require.NoError(t, err)

	counts, err := s.Store.FindAccountDataCounts(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 0, counts.Expenses)
	require.Equal(t, 0, counts.RecurrentExpenses)
	require.Equal(t, 0, counts.MacroEntries)
	require.Equal(t, 0, counts.MacroGoals)
	require.Equal(t, 0, counts.Foods)
	require.Equal(t, 0, counts.MoodEntries)
	require.Equal(t, 0, counts.Tags)

	otherCounts, err := s.Store.FindAccountDataCounts(ctx, otherUser.ID)
	require.NoError(t, err)
	require.Equal(t, 1, otherCounts.Expenses)
	require.Equal(t, 1, otherCounts.RecurrentExpenses)
	require.Equal(t, 1, otherCounts.MacroEntries)
	require.Equal(t, 1, otherCounts.MacroGoals)
	require.Equal(t, 1, otherCounts.Foods)
	require.Equal(t, 1, otherCounts.MoodEntries)
	// Two tags per seed: one from the expense, one from the mood entry.
	require.Equal(t, 2, otherCounts.Tags)
}
