package logic_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateRecurrentExpense(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_user_1",
		Email:        "recurrent_user_1@example.com",
		PasswordHash: []byte("recurrent_user_hash_1"),
	})
	category := s.CreateCategory(t, "recurrent category 1")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_recurrent_expense",
			fn: func(t *testing.T) {
				recurrentExpense, err := s.Store.CreateRecurrentExpense(
					ctx,
					user.ID,
					newRecurrentExpenseParams(category.ID, "recurrent description 1", 2000, 1),
				)
				require.NoError(t, err)
				require.Positive(t, recurrentExpense.ID)
				require.Equal(t, user.ID, recurrentExpense.UserID)
			},
		},
		{
			name: "should_fail_validation_for_invalid_params",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateRecurrentExpense(
					ctx,
					user.ID,
					newRecurrentExpenseParams(0, "no", 0, 0),
				)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindRecurrentExpense(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_user_2",
		Email:        "recurrent_user_2@example.com",
		PasswordHash: []byte("recurrent_user_hash_2"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_user_3",
		Email:        "recurrent_user_3@example.com",
		PasswordHash: []byte("recurrent_user_hash_3"),
	})
	category := s.CreateCategory(t, "recurrent category 2")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_recurrent_expense_for_owner",
			fn: func(t *testing.T) {
				recurrentExpense := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(category.ID, "recurrent description 2", 2100, 2),
				)

				foundRecurrentExpense, err := s.Store.FindRecurrentExpense(ctx, recurrentExpense.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, recurrentExpense.ID, foundRecurrentExpense.ID)
			},
		},
		{
			name: "should_fail_when_recurrent_expense_does_not_belong_to_user",
			fn: func(t *testing.T) {
				recurrentExpense := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(category.ID, "recurrent description 3", 2200, 3),
				)

				_, err := s.Store.FindRecurrentExpense(ctx, recurrentExpense.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindRecurrentExpenses(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_user_4",
		Email:        "recurrent_user_4@example.com",
		PasswordHash: []byte("recurrent_user_hash_4"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_user_5",
		Email:        "recurrent_user_5@example.com",
		PasswordHash: []byte("recurrent_user_hash_5"),
	})
	category := s.CreateCategory(t, "recurrent category 3")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_recurrent_expenses_for_filtered_user",
			fn: func(t *testing.T) {
				recurrentExpenseOne := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(category.ID, "recurrent description 4", 2300, 1),
				)
				recurrentExpenseTwo := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(category.ID, "recurrent description 5", 2400, 1),
				)
				s.CreateRecurrentExpense(
					t,
					otherUser.ID,
					newRecurrentExpenseParams(category.ID, "recurrent description 6", 2500, 1),
				)

				recurrentExpenses, err := s.Store.FindRecurrentExpenses(ctx, repo.QueryOptions{
					Filters: repo.Filters{
						FilterFields: []repo.FilterField{
							{Name: "user_id", Value: user.ID, Operator: "="},
						},
					},
					Sorting: repo.Sorting{Field: "id", Order: "ASC"},
				})
				require.NoError(t, err)
				require.Len(t, recurrentExpenses, 2)
				require.Equal(t, recurrentExpenseOne.ID, recurrentExpenses[0].ID)
				require.Equal(t, recurrentExpenseTwo.ID, recurrentExpenses[1].ID)
			},
		},
		{
			name: "should_fail_with_invalid_sort_field",
			fn: func(t *testing.T) {
				_, err := s.Store.FindRecurrentExpenses(ctx, repo.QueryOptions{
					Sorting: repo.Sorting{Field: "invalid_field", Order: "ASC"},
				})
				require.ErrorIs(t, err, repo.ErrInvalidField)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateRecurrentExpense(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_user_6",
		Email:        "recurrent_user_6@example.com",
		PasswordHash: []byte("recurrent_user_hash_6"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_user_7",
		Email:        "recurrent_user_7@example.com",
		PasswordHash: []byte("recurrent_user_hash_7"),
	})
	categoryOne := s.CreateCategory(t, "recurrent category 4")
	categoryTwo := s.CreateCategory(t, "recurrent category 5")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_recurrent_expense",
			fn: func(t *testing.T) {
				recurrentExpense := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(categoryOne.ID, "recurrent description 7", 2600, 1),
				)

				updatedRecurrentExpense, err := s.Store.UpdateRecurrentExpense(
					ctx,
					recurrentExpense.ID,
					user.ID,
					newRecurrentExpenseParams(categoryTwo.ID, "recurrent description 7 updated", 2700, 2),
				)
				require.NoError(t, err)
				require.Equal(t, categoryTwo.ID, updatedRecurrentExpense.CategoryID)
				require.Equal(t, uint(2), updatedRecurrentExpense.Period)
			},
		},
		{
			name: "should_fail_when_recurrent_expense_does_not_belong_to_user",
			fn: func(t *testing.T) {
				recurrentExpense := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(categoryOne.ID, "recurrent description 8", 2800, 1),
				)

				_, err := s.Store.UpdateRecurrentExpense(
					ctx,
					recurrentExpense.ID,
					otherUser.ID,
					newRecurrentExpenseParams(categoryTwo.ID, "recurrent description 8 updated", 2900, 2),
				)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
		{
			name: "should_fail_validation_for_invalid_params",
			fn: func(t *testing.T) {
				recurrentExpense := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(categoryOne.ID, "recurrent description 9", 3000, 1),
				)

				_, err := s.Store.UpdateRecurrentExpense(
					ctx,
					recurrentExpense.ID,
					user.ID,
					newRecurrentExpenseParams(0, "no", 0, 0),
				)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteRecurrentExpense(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_user_8",
		Email:        "recurrent_user_8@example.com",
		PasswordHash: []byte("recurrent_user_hash_8"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_user_9",
		Email:        "recurrent_user_9@example.com",
		PasswordHash: []byte("recurrent_user_hash_9"),
	})
	category := s.CreateCategory(t, "recurrent category 6")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_recurrent_expense_for_owner",
			fn: func(t *testing.T) {
				recurrentExpense := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(category.ID, "recurrent description 10", 3100, 1),
				)

				deletedID, err := s.Store.DeleteRecurrentExpense(ctx, recurrentExpense.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, recurrentExpense.ID, deletedID)
			},
		},
		{
			name: "should_fail_when_deleting_recurrent_expense_of_another_user",
			fn: func(t *testing.T) {
				recurrentExpense := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(category.ID, "recurrent description 11", 3200, 1),
				)

				_, err := s.Store.DeleteRecurrentExpense(ctx, recurrentExpense.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestCopyDueRecurrentExpenses(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_user_copy_1",
		Email:        "recurrent_user_copy_1@example.com",
		PasswordHash: []byte("recurrent_user_copy_hash_1"),
	})
	category := s.CreateCategory(t, "recurrent category copy 1")
	now := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	expenseDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Unix()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_copy_when_last_copy_is_null",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(category.ID, "copy null last 1", 5000, 1),
				)

				copied, err := s.Store.CopyDueRecurrentExpenses(ctx, now)
				require.NoError(t, err)
				require.GreaterOrEqual(t, copied, 1)

				expenses, err := s.Store.FindExpenses(ctx, repo.QueryOptions{
					Filters: repo.Filters{
						FilterFields: []repo.FilterField{
							{Name: "user_id", Value: user.ID, Operator: "="},
							{Name: "description", Value: re.Description, Operator: "="},
						},
						Connector: "AND",
					},
				})
				require.NoError(t, err)
				require.Len(t, expenses, 1)
				require.Equal(t, expenseDate, expenses[0].Date)

				updated, err := s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.NotNil(t, updated.LastCopyCreatedAt)
				require.Equal(t, expenseDate, *updated.LastCopyCreatedAt)
			},
		},
		{
			name: "should_copy_when_period_has_elapsed",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(category.ID, "copy elapsed 1", 6000, 1),
				)
				twoMonthsAgo := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
				s.SetRecurrentExpenseLastCopy(t, re, twoMonthsAgo)

				copied, err := s.Store.CopyDueRecurrentExpenses(ctx, now)
				require.NoError(t, err)
				require.GreaterOrEqual(t, copied, 1)

				updated, err := s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.NotNil(t, updated.LastCopyCreatedAt)
				require.Equal(t, expenseDate, *updated.LastCopyCreatedAt)
			},
		},
		{
			name: "should_not_copy_when_period_not_elapsed",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(
					t,
					user.ID,
					newRecurrentExpenseParams(category.ID, "copy not elapsed 1", 7000, 3),
				)
				oneMonthAgo := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC).Unix()
				s.SetRecurrentExpenseLastCopy(t, re, oneMonthAgo)

				_, err := s.Store.CopyDueRecurrentExpenses(ctx, now)
				require.NoError(t, err)

				updated, err := s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.NotNil(t, updated.LastCopyCreatedAt)
				require.Equal(t, oneMonthAgo, *updated.LastCopyCreatedAt)
			},
		},
		{
			name: "should_return_zero_when_no_due_expenses",
			fn: func(t *testing.T) {
				// Use a date far enough in the past that nothing is due
				pastNow := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
				copied, err := s.Store.CopyDueRecurrentExpenses(ctx, pastNow)
				require.NoError(t, err)
				require.Equal(t, 0, copied)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func newRecurrentExpenseParams(
	categoryID int,
	description string,
	amount uint64,
	period uint,
) logic.RecurrentExpenseParams {
	return logic.RecurrentExpenseParams{
		ExpenseBaseParams: logic.ExpenseBaseParams{
			CategoryID:  categoryID,
			Description: description,
			Amount:      amount,
		},
		Period: period,
	}
}

func TestRecurrentExpenseTags(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_tag_user_1",
		Email:        "recurrent_tag_user_1@example.com",
		PasswordHash: []byte("recurrent_tag_user_hash_1"),
	})
	category := s.CreateCategory(t, "recurrent tag category 1")
	now := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_attach_tags_on_create",
			fn: func(t *testing.T) {
				params := newRecurrentExpenseParams(category.ID, "recurrent tags create 1", 4100, 1)
				params.Tags = []string{"Rent", "fixed"}

				re, err := s.Store.CreateRecurrentExpense(ctx, user.ID, params)
				require.NoError(t, err)

				tags, err := s.Store.FindRecurrentExpenseTags(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, []string{"fixed", "rent"}, logic.ExtractTagNames(tags))
			},
		},
		{
			name: "should_replace_tags_on_update",
			fn: func(t *testing.T) {
				params := newRecurrentExpenseParams(category.ID, "recurrent tags update 1", 4200, 1)
				params.Tags = []string{"old"}

				re, err := s.Store.CreateRecurrentExpense(ctx, user.ID, params)
				require.NoError(t, err)

				params.Tags = []string{"new"}
				_, err = s.Store.UpdateRecurrentExpense(ctx, re.ID, user.ID, params)
				require.NoError(t, err)

				tags, err := s.Store.FindRecurrentExpenseTags(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, []string{"new"}, logic.ExtractTagNames(tags))
			},
		},
		{
			name: "should_delete_taggings_with_the_recurrent_expense",
			fn: func(t *testing.T) {
				params := newRecurrentExpenseParams(category.ID, "recurrent tags delete 1", 4300, 1)
				params.Tags = []string{"doomed"}

				re, err := s.Store.CreateRecurrentExpense(ctx, user.ID, params)
				require.NoError(t, err)

				_, err = s.Store.DeleteRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)

				count, err := s.Queries.CountTaggingsByTarget(ctx, repo.TaggableTypeRecurrentExpense, re.ID)
				require.NoError(t, err)
				require.Zero(t, count)
			},
		},
		{
			name: "should_copy_tags_onto_the_generated_expense",
			fn: func(t *testing.T) {
				params := newRecurrentExpenseParams(category.ID, "recurrent tags copy 1", 4400, 1)
				params.Tags = []string{"subscription", "monthly"}

				re, err := s.Store.CreateRecurrentExpense(ctx, user.ID, params)
				require.NoError(t, err)

				_, err = s.Store.CopyDueRecurrentExpenses(ctx, now)
				require.NoError(t, err)

				expenses, err := s.Store.FindExpenses(ctx, repo.QueryOptions{
					Filters: repo.Filters{
						FilterFields: []repo.FilterField{
							{Name: "user_id", Value: user.ID, Operator: "="},
							{Name: "description", Value: re.Description, Operator: "="},
						},
						Connector: "AND",
					},
				})
				require.NoError(t, err)
				require.Len(t, expenses, 1)

				expenseTags, err := s.Store.FindExpenseTags(ctx, expenses[0].ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, []string{"monthly", "subscription"}, logic.ExtractTagNames(expenseTags))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestRecurrentExpenseOccurrenceLimit(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "recurrent_limit_user_1",
		Email:        "recurrent_limit_user_1@example.com",
		PasswordHash: []byte("recurrent_limit_user_hash_1"),
	})
	category := s.CreateCategory(t, "recurrent category limit 1")
	march := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	april := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	may := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)

	newLimitedParams := func(description string, limit uint) logic.RecurrentExpenseParams {
		params := newRecurrentExpenseParams(category.ID, description, 5000, 1)
		params.OccurrenceLimit = limit

		return params
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_archive_once_the_limit_is_reached",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(t, user.ID, newLimitedParams("limit archive 1", 2))

				_, err := s.Store.CopyDueRecurrentExpenses(ctx, march)
				require.NoError(t, err)

				updated, err := s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, uint(1), updated.OccurrenceCount)
				require.Nil(t, updated.ArchivedAt)

				_, err = s.Store.CopyDueRecurrentExpenses(ctx, april)
				require.NoError(t, err)

				updated, err = s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, uint(2), updated.OccurrenceCount)
				require.NotNil(t, updated.ArchivedAt)
			},
		},
		{
			name: "should_not_copy_an_archived_recurrent_expense",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(t, user.ID, newLimitedParams("limit skip 1", 1))

				_, err := s.Store.CopyDueRecurrentExpenses(ctx, march)
				require.NoError(t, err)

				_, err = s.Store.CopyDueRecurrentExpenses(ctx, may)
				require.NoError(t, err)

				updated, err := s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, uint(1), updated.OccurrenceCount)

				expenses, err := s.Store.FindExpenses(ctx, repo.QueryOptions{
					Filters: repo.Filters{
						FilterFields: []repo.FilterField{
							{Name: "user_id", Value: user.ID, Operator: "="},
							{Name: "description", Value: re.Description, Operator: "="},
						},
						Connector: "AND",
					},
				})
				require.NoError(t, err)
				require.Len(t, expenses, 1)
			},
		},
		{
			name: "should_never_archive_when_the_limit_is_zero",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(t, user.ID, newLimitedParams("limit unlimited 1", 0))

				_, err := s.Store.CopyDueRecurrentExpenses(ctx, march)
				require.NoError(t, err)
				_, err = s.Store.CopyDueRecurrentExpenses(ctx, april)
				require.NoError(t, err)

				updated, err := s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, uint(2), updated.OccurrenceCount)
				require.Nil(t, updated.ArchivedAt)
			},
		},
		{
			name: "should_reset_the_count_when_unarchived",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(t, user.ID, newLimitedParams("limit unarchive 1", 1))

				_, err := s.Store.CopyDueRecurrentExpenses(ctx, march)
				require.NoError(t, err)

				archived, err := s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.NotNil(t, archived.ArchivedAt)

				unarchived, err := s.Store.UnarchiveRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.Nil(t, unarchived.ArchivedAt)
				require.Equal(t, uint(0), unarchived.OccurrenceCount)

				_, err = s.Store.CopyDueRecurrentExpenses(ctx, april)
				require.NoError(t, err)

				updated, err := s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, uint(1), updated.OccurrenceCount)
				require.NotNil(t, updated.ArchivedAt)
			},
		},
		{
			name: "should_archive_when_the_limit_is_lowered_to_the_count",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(t, user.ID, newLimitedParams("limit lowered 1", 5))

				_, err := s.Store.CopyDueRecurrentExpenses(ctx, march)
				require.NoError(t, err)
				_, err = s.Store.CopyDueRecurrentExpenses(ctx, april)
				require.NoError(t, err)

				lowered, err := s.Store.UpdateRecurrentExpense(
					ctx,
					re.ID,
					user.ID,
					newLimitedParams("limit lowered 1", 2),
				)
				require.NoError(t, err)
				require.Equal(t, uint(2), lowered.OccurrenceCount)
				require.NotNil(t, lowered.ArchivedAt)

				_, err = s.Store.CopyDueRecurrentExpenses(ctx, may)
				require.NoError(t, err)

				updated, err := s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, uint(2), updated.OccurrenceCount)
			},
		},
		{
			name: "should_not_archive_when_the_limit_is_lowered_above_the_count",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(t, user.ID, newLimitedParams("limit lowered 2", 5))

				_, err := s.Store.CopyDueRecurrentExpenses(ctx, march)
				require.NoError(t, err)

				lowered, err := s.Store.UpdateRecurrentExpense(
					ctx,
					re.ID,
					user.ID,
					newLimitedParams("limit lowered 2", 3),
				)
				require.NoError(t, err)
				require.Equal(t, uint(1), lowered.OccurrenceCount)
				require.Nil(t, lowered.ArchivedAt)
			},
		},
		{
			name: "should_keep_an_archived_row_archived_when_the_limit_is_raised",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(t, user.ID, newLimitedParams("limit raised 1", 1))

				_, err := s.Store.CopyDueRecurrentExpenses(ctx, march)
				require.NoError(t, err)

				raised, err := s.Store.UpdateRecurrentExpense(
					ctx,
					re.ID,
					user.ID,
					newLimitedParams("limit raised 1", 4),
				)
				require.NoError(t, err)
				require.NotNil(t, raised.ArchivedAt)
			},
		},
		{
			name: "should_not_unarchive_an_active_recurrent_expense",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(t, user.ID, newLimitedParams("limit active unarchive 1", 5))

				_, err := s.Store.CopyDueRecurrentExpenses(ctx, march)
				require.NoError(t, err)

				_, err = s.Store.UnarchiveRecurrentExpense(ctx, re.ID, user.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)

				untouched, err := s.Store.FindRecurrentExpense(ctx, re.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, uint(1), untouched.OccurrenceCount)
			},
		},
		{
			name: "should_not_unarchive_another_users_recurrent_expense",
			fn: func(t *testing.T) {
				other := s.CreateUser(t, repo.InsertUserParams{
					Username:     "recurrent_limit_user_2",
					Email:        "recurrent_limit_user_2@example.com",
					PasswordHash: []byte("recurrent_limit_user_hash_2"),
				})
				re := s.CreateRecurrentExpense(t, user.ID, newLimitedParams("limit ownership 1", 1))

				_, err := s.Store.UnarchiveRecurrentExpense(ctx, re.ID, other.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
		{
			name: "should_keep_the_count_when_the_recurrent_expense_is_edited",
			fn: func(t *testing.T) {
				re := s.CreateRecurrentExpense(t, user.ID, newLimitedParams("limit edit 1", 3))

				_, err := s.Store.CopyDueRecurrentExpenses(ctx, march)
				require.NoError(t, err)

				params := newLimitedParams("limit edit 1 renamed", 5)
				updated, err := s.Store.UpdateRecurrentExpense(ctx, re.ID, user.ID, params)
				require.NoError(t, err)
				require.Equal(t, uint(1), updated.OccurrenceCount)
				require.Equal(t, uint(5), updated.OccurrenceLimit)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
