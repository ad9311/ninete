package logic_test

import (
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestFindExpenses(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "findexpensesuser",
		Email:                "findexpenses@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Find Expenses Category")

	date := time.Now().UTC().Unix()
	expenseOne := f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "Morning coffee",
		Amount:      500,
		Date:        date,
	})
	expenseTwo := f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "Lunch",
		Amount:      1500,
		Date:        date + 10,
	})

	otherUser := f.User(t, logic.SignUpParams{
		Username:             "findexpensesother",
		Email:                "findexpensesother@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	f.Expense(t, repo.InsertExpenseParams{
		UserID:      otherUser.ID,
		CategoryID:  category.ID,
		Description: "Other user expense",
		Amount:      100,
		Date:        date,
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_find_expenses",
			func(t *testing.T) {
				opts := repo.QueryOptions{
					Filters: repo.Filters{
						FilterFields: []repo.FilterField{
							{Name: "user_id", Value: user.ID, Operator: "="},
						},
					},
					Sorting: repo.Sorting{Field: "id", Order: "ASC"},
				}
				expenses, err := f.Store.FindExpenses(ctx, opts)
				require.NoError(t, err)
				require.Len(t, expenses, 2)
				require.Equal(t, expenseOne.Description, expenses[0].Description)
				require.Equal(t, expenseTwo.Description, expenses[1].Description)
				require.Equal(t, user.ID, expenses[0].UserID)
				require.Equal(t, user.ID, expenses[1].UserID)
			},
		},
		{
			"should_fail_invalid_sorting_field",
			func(t *testing.T) {
				opts := repo.QueryOptions{
					Sorting: repo.Sorting{Field: "invalid", Order: "ASC"},
				}
				_, err := f.Store.FindExpenses(ctx, opts)
				require.ErrorIs(t, err, repo.ErrInvalidField)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestCountExpenses(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "countexpensesuser",
		Email:                "countexpenses@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Count Expenses Category")

	date := time.Now().UTC().Unix()
	f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "Breakfast",
		Amount:      700,
		Date:        date,
	})
	f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "Dinner",
		Amount:      2000,
		Date:        date + 20,
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_count_user_expenses",
			func(t *testing.T) {
				filters := repo.Filters{
					FilterFields: []repo.FilterField{
						{Name: "user_id", Value: user.ID, Operator: "="},
					},
				}
				count, err := f.Store.CountExpenses(ctx, filters)
				require.NoError(t, err)
				require.Equal(t, 2, count)
			},
		},
		{
			"should_fail_invalid_operator",
			func(t *testing.T) {
				filters := repo.Filters{
					FilterFields: []repo.FilterField{
						{Name: "user_id", Value: user.ID, Operator: "??"},
					},
				}
				_, err := f.Store.CountExpenses(ctx, filters)
				require.ErrorIs(t, err, repo.ErrInvalidOperator)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "findexpenseuser",
		Email:                "findexpense@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Find Expense Category")

	expense := f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "Groceries",
		Amount:      3000,
		Date:        time.Now().UTC().Unix(),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_find_expense",
			func(t *testing.T) {
				e, err := f.Store.FindExpense(ctx, expense.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, expense.ID, e.ID)
				require.Equal(t, expense.Description, e.Description)
				require.Equal(t, expense.Amount, e.Amount)
				require.Equal(t, user.ID, e.UserID)
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				_, err := f.Store.FindExpense(ctx, -1, user.ID)
				require.ErrorIs(t, err, logic.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestCreateExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "createexpenseuser",
		Email:                "createexpense@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Create Expense Category")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_create_expense",
			func(t *testing.T) {
				params := repo.InsertExpenseParams{
					UserID:      user.ID,
					CategoryID:  category.ID,
					Description: "Rent",
					Amount:      120000,
					Date:        time.Now().UTC().Unix(),
				}
				expense := f.Expense(t, params)
				require.Equal(t, params.Description, expense.Description)
				require.Equal(t, params.Amount, expense.Amount)
				require.Equal(t, user.ID, expense.UserID)
				require.Equal(t, category.ID, expense.CategoryID)
			},
		},
		{
			"should_fail_validation",
			func(t *testing.T) {
				params := logic.ExpenseParams{
					CategoryID:  category.ID,
					Description: "ab",
					Amount:      0,
					Date:        0,
				}
				_, err := f.Store.CreateExpense(ctx, user.ID, params)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Description:min]")
				require.Contains(t, err.Error(), "[Amount:required]")
				require.Contains(t, err.Error(), "[Date:required]")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "updateexpenseuser",
		Email:                "updateexpense@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Update Expense Category")
	expense := f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "Subscription",
		Amount:      900,
		Date:        time.Now().UTC().Unix(),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_update_expense",
			func(t *testing.T) {
				params := logic.ExpenseParams{
					CategoryID:  category.ID,
					Description: "Subscription upgraded",
					Amount:      1500,
					Date:        expense.Date + 50,
				}
				updated, err := f.Store.UpdateExpense(ctx, expense.ID, params)
				require.NoError(t, err)
				require.Equal(t, expense.ID, updated.ID)
				require.Equal(t, params.Description, updated.Description)
				require.Equal(t, params.Amount, updated.Amount)
				require.Equal(t, params.Date, updated.Date)
			},
		},
		{
			"should_fail_validation",
			func(t *testing.T) {
				params := logic.ExpenseParams{
					CategoryID:  category.ID,
					Description: "",
					Amount:      0,
					Date:        0,
				}
				_, err := f.Store.UpdateExpense(ctx, expense.ID, params)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Description:required]")
				require.Contains(t, err.Error(), "[Amount:required]")
				require.Contains(t, err.Error(), "[Date:required]")
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				params := logic.ExpenseParams{
					CategoryID:  category.ID,
					Description: "Missing expense",
					Amount:      100,
					Date:        expense.Date,
				}
				_, err := f.Store.UpdateExpense(ctx, -1, params)
				require.ErrorIs(t, err, logic.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "deleteexpenseuser",
		Email:                "deleteexpense@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Delete Expense Category")
	expense := f.Expense(t, repo.InsertExpenseParams{
		UserID:      user.ID,
		CategoryID:  category.ID,
		Description: "To be deleted",
		Amount:      400,
		Date:        time.Now().UTC().Unix(),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_delete_expense",
			func(t *testing.T) {
				deletedID, err := f.Store.DeleteExpense(ctx, expense.ID)
				require.NoError(t, err)
				require.Equal(t, expense.ID, deletedID)

				_, findErr := f.Store.FindExpense(ctx, expense.ID, user.ID)
				require.ErrorIs(t, findErr, logic.ErrNotFound)
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				_, err := f.Store.DeleteExpense(ctx, -1)
				require.ErrorIs(t, err, logic.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
