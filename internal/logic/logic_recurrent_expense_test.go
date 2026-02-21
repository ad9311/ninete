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
