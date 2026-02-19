package logic_test

import (
	"database/sql"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateExpense(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_1",
		Email:        "expense_user_1@example.com",
		PasswordHash: []byte("expense_user_hash_1"),
	})
	category := s.CreateCategory(t, "expense category 1")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_expense_with_normalized_tags",
			fn: func(t *testing.T) {
				expense, err := s.Store.CreateExpense(
					ctx,
					user.ID,
					newExpenseParams(
						category.ID,
						"expense description 1",
						550,
						1735689600,
						[]string{"TAG_A_1", " tag_a_1 ", "tag_b_1"},
					),
				)
				require.NoError(t, err)
				require.Positive(t, expense.ID)
				require.Equal(t, user.ID, expense.UserID)
				require.Equal(t, category.ID, expense.CategoryID)

				tags, err := s.Store.FindExpenseTags(ctx, expense.ID, user.ID)
				require.NoError(t, err)
				require.Len(t, tags, 2)
				require.Equal(t, "tag_a_1", tags[0].Name)
				require.Equal(t, "tag_b_1", tags[1].Name)
			},
		},
		{
			name: "should_fail_validation_for_invalid_params",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateExpense(ctx, user.ID, newExpenseParams(0, "no", 0, 0, nil))
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindExpenses(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_2",
		Email:        "expense_user_2@example.com",
		PasswordHash: []byte("expense_user_hash_2"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_3",
		Email:        "expense_user_3@example.com",
		PasswordHash: []byte("expense_user_hash_3"),
	})
	category := s.CreateCategory(t, "expense category 2")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_expenses_for_filtered_user",
			fn: func(t *testing.T) {
				expenseOne := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(category.ID, "expense description 2", 100, 1735776000, nil),
				)
				expenseTwo := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(category.ID, "expense description 3", 200, 1735862400, nil),
				)
				s.CreateExpense(t, otherUser.ID, newExpenseParams(category.ID, "expense description 4", 300, 1735948800, nil))

				expenses, err := s.Store.FindExpenses(ctx, repo.QueryOptions{
					Filters: repo.Filters{
						FilterFields: []repo.FilterField{
							{Name: "user_id", Value: user.ID, Operator: "="},
						},
					},
					Sorting: repo.Sorting{
						Field: "date",
						Order: "ASC",
					},
				})
				require.NoError(t, err)
				require.Len(t, expenses, 2)
				require.Equal(t, expenseOne.ID, expenses[0].ID)
				require.Equal(t, expenseTwo.ID, expenses[1].ID)
			},
		},
		{
			name: "should_fail_with_invalid_sort_field",
			fn: func(t *testing.T) {
				_, err := s.Store.FindExpenses(ctx, repo.QueryOptions{
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

func TestCountExpenses(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_4",
		Email:        "expense_user_4@example.com",
		PasswordHash: []byte("expense_user_hash_4"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_5",
		Email:        "expense_user_5@example.com",
		PasswordHash: []byte("expense_user_hash_5"),
	})
	category := s.CreateCategory(t, "expense category 3")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_count_filtered_expenses",
			fn: func(t *testing.T) {
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "expense description 5", 100, 1736035200, nil))
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "expense description 6", 200, 1736121600, nil))
				s.CreateExpense(t, otherUser.ID, newExpenseParams(category.ID, "expense description 7", 300, 1736208000, nil))

				count, err := s.Store.CountExpenses(ctx, repo.Filters{
					FilterFields: []repo.FilterField{
						{Name: "user_id", Value: user.ID, Operator: "="},
					},
				})
				require.NoError(t, err)
				require.Equal(t, 2, count)
			},
		},
		{
			name: "should_fail_with_invalid_filter_operator",
			fn: func(t *testing.T) {
				_, err := s.Store.CountExpenses(ctx, repo.Filters{
					FilterFields: []repo.FilterField{
						{Name: "user_id", Value: user.ID, Operator: "!="},
					},
				})
				require.ErrorIs(t, err, repo.ErrInvalidOperator)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindExpense(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_6",
		Email:        "expense_user_6@example.com",
		PasswordHash: []byte("expense_user_hash_6"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_7",
		Email:        "expense_user_7@example.com",
		PasswordHash: []byte("expense_user_hash_7"),
	})
	category := s.CreateCategory(t, "expense category 4")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_expense_by_id_for_owner",
			fn: func(t *testing.T) {
				expense := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(category.ID, "expense description 8", 700, 1736294400, nil),
				)
				foundExpense, err := s.Store.FindExpense(ctx, expense.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, expense.ID, foundExpense.ID)
			},
		},
		{
			name: "should_fail_when_expense_does_not_belong_to_user",
			fn: func(t *testing.T) {
				expense := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(category.ID, "expense description 9", 800, 1736380800, nil),
				)
				_, err := s.Store.FindExpense(ctx, expense.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindExpenseTagsAndRows(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_8",
		Email:        "expense_user_8@example.com",
		PasswordHash: []byte("expense_user_hash_8"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_9",
		Email:        "expense_user_9@example.com",
		PasswordHash: []byte("expense_user_hash_9"),
	})
	category := s.CreateCategory(t, "expense category 5")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_expense_tags_and_rows",
			fn: func(t *testing.T) {
				expenseOne := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(category.ID, "expense description 10", 900, 1736467200, []string{"tag_c_1", "tag_a_1"}),
				)
				expenseTwo := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(category.ID, "expense description 11", 1000, 1736553600, []string{"tag_b_1"}),
				)

				tags, err := s.Store.FindExpenseTags(ctx, expenseOne.ID, user.ID)
				require.NoError(t, err)
				require.Len(t, tags, 2)
				require.Equal(t, "tag_a_1", tags[0].Name)
				require.Equal(t, "tag_c_1", tags[1].Name)

				rows, err := s.Store.FindExpenseTagRows(ctx, []int{expenseOne.ID, expenseTwo.ID}, user.ID)
				require.NoError(t, err)
				require.Len(t, rows, 3)
				require.Equal(t, expenseOne.ID, rows[0].ExpenseID)
				require.Equal(t, "tag_a_1", rows[0].TagName)
				require.Equal(t, expenseTwo.ID, rows[2].ExpenseID)
				require.Equal(t, "tag_b_1", rows[2].TagName)
			},
		},
		{
			name: "should_return_empty_tags_for_non_owner_and_empty_rows_for_empty_ids",
			fn: func(t *testing.T) {
				expense := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(category.ID, "expense description 12", 1100, 1736640000, []string{"tag_d_1"}),
				)

				tags, err := s.Store.FindExpenseTags(ctx, expense.ID, otherUser.ID)
				require.NoError(t, err)
				require.Empty(t, tags)

				rows, err := s.Store.FindExpenseTagRows(ctx, []int{}, user.ID)
				require.NoError(t, err)
				require.Empty(t, rows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateExpense(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_10",
		Email:        "expense_user_10@example.com",
		PasswordHash: []byte("expense_user_hash_10"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_11",
		Email:        "expense_user_11@example.com",
		PasswordHash: []byte("expense_user_hash_11"),
	})
	categoryOne := s.CreateCategory(t, "expense category 6")
	categoryTwo := s.CreateCategory(t, "expense category 7")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_expense_and_replace_tags",
			fn: func(t *testing.T) {
				expense := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(categoryOne.ID, "expense description 13", 1200, 1736726400, []string{"old_tag_1"}),
				)

				updatedExpense, err := s.Store.UpdateExpense(
					ctx,
					expense.ID,
					user.ID,
					newExpenseParams(categoryTwo.ID, "expense description 13 updated", 1300, 1736812800, []string{"new_tag_1"}),
				)
				require.NoError(t, err)
				require.Equal(t, categoryTwo.ID, updatedExpense.CategoryID)
				require.Equal(t, "expense description 13 updated", updatedExpense.Description)

				tags, err := s.Store.FindExpenseTags(ctx, expense.ID, user.ID)
				require.NoError(t, err)
				require.Len(t, tags, 1)
				require.Equal(t, "new_tag_1", tags[0].Name)
			},
		},
		{
			name: "should_fail_when_expense_does_not_belong_to_user",
			fn: func(t *testing.T) {
				expense := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(categoryOne.ID, "expense description 14", 1400, 1736899200, nil),
				)
				_, err := s.Store.UpdateExpense(
					ctx,
					expense.ID,
					otherUser.ID,
					newExpenseParams(categoryTwo.ID, "expense description 14 updated", 1500, 1736985600, nil),
				)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
		{
			name: "should_fail_validation_for_invalid_params",
			fn: func(t *testing.T) {
				expense := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(categoryOne.ID, "expense description 15", 1600, 1737072000, nil),
				)
				_, err := s.Store.UpdateExpense(ctx, expense.ID, user.ID, newExpenseParams(0, "no", 0, 0, nil))
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteExpense(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_12",
		Email:        "expense_user_12@example.com",
		PasswordHash: []byte("expense_user_hash_12"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "expense_user_13",
		Email:        "expense_user_13@example.com",
		PasswordHash: []byte("expense_user_hash_13"),
	})
	category := s.CreateCategory(t, "expense category 8")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_expense_for_owner",
			fn: func(t *testing.T) {
				expense := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(category.ID, "expense description 16", 1700, 1737158400, nil),
				)
				deletedID, err := s.Store.DeleteExpense(ctx, expense.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, expense.ID, deletedID)
			},
		},
		{
			name: "should_fail_when_deleting_expense_of_another_user",
			fn: func(t *testing.T) {
				expense := s.CreateExpense(
					t,
					user.ID,
					newExpenseParams(category.ID, "expense description 17", 1800, 1737244800, nil),
				)
				_, err := s.Store.DeleteExpense(ctx, expense.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestExtractTagNames(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_return_tag_names_in_order",
			fn: func(t *testing.T) {
				tagNames := logic.ExtractTagNames([]repo.Tag{
					{Name: "tag_a_1"},
					{Name: "tag_b_1"},
					{Name: "tag_c_1"},
				})
				require.Equal(t, []string{"tag_a_1", "tag_b_1", "tag_c_1"}, tagNames)
			},
		},
		{
			name: "should_return_empty_slice_for_empty_tags",
			fn: func(t *testing.T) {
				tagNames := logic.ExtractTagNames([]repo.Tag{})
				require.Empty(t, tagNames)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func newExpenseParams(
	categoryID int,
	description string,
	amount uint64,
	date int64,
	tags []string,
) logic.ExpenseParams {
	return logic.ExpenseParams{
		ExpenseBaseParams: logic.ExpenseBaseParams{
			CategoryID:  categoryID,
			Description: description,
			Amount:      amount,
		},
		Date: date,
		Tags: tags,
	}
}
