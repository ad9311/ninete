package logic_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestFindRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "findrecurrentuser",
		Email:                "findrecurrent@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Find Recurrent Category")

	recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Monthly internet",
		Amount:      4500,
		Period:      1,
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_find_recurrent_expense",
			func(t *testing.T) {
				re, err := f.Store.FindRecurrentExpense(ctx, recurrent.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, recurrent.ID, re.ID)
				require.Equal(t, recurrent.Description, re.Description)
				require.Equal(t, recurrent.Amount, re.Amount)
				require.Equal(t, user.ID, re.UserID)
				require.Equal(t, category.ID, re.CategoryID)
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				_, err := f.Store.FindRecurrentExpense(ctx, -1, user.ID)
				require.ErrorIs(t, err, logic.ErrNotFound)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestCreateRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "createrecurrentuser",
		Email:                "createrecurrent@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Create Recurrent Category")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_create_recurrent_expense",
			func(t *testing.T) {
				params := logic.RecurrentExpenseParams{
					CategoryID:  category.ID,
					Description: "Gym membership",
					Amount:      2500,
					Period:      1,
				}
				recurrent, err := f.Store.CreateRecurrentExpense(ctx, user.ID, params)
				require.NoError(t, err)
				require.Equal(t, params.Description, recurrent.Description)
				require.Equal(t, params.Amount, recurrent.Amount)
				require.Equal(t, params.Period, recurrent.Period)
				require.Equal(t, user.ID, recurrent.UserID)
				require.Equal(t, category.ID, recurrent.CategoryID)
			},
		},
		{
			"should_fail_validation",
			func(t *testing.T) {
				params := logic.RecurrentExpenseParams{
					CategoryID:  0,
					Description: "ab",
					Amount:      0,
					Period:      0,
				}
				_, err := f.Store.CreateRecurrentExpense(ctx, user.ID, params)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[CategoryID:required]")
				require.Contains(t, err.Error(), "[Description:min]")
				require.Contains(t, err.Error(), "[Amount:required]")
				require.Contains(t, err.Error(), "[Period:required]")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateRecurrentExpense(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "updaterecurrentuser",
		Email:                "updaterecurrent@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Update Recurrent Category")
	recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Streaming",
		Amount:      1200,
		Period:      1,
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_update_recurrent_expense",
			func(t *testing.T) {
				lastCopy := sql.NullInt64{Valid: true, Int64: time.Now().Unix()}
				params := repo.UpdateRecurrentExpenseParams{
					ID:                recurrent.ID,
					UserID:            user.ID,
					Description:       "Streaming premium",
					Amount:            1800,
					Period:            2,
					LastCopyCreatedAt: lastCopy,
				}
				updated, err := f.Store.UpdateRecurrentExpense(ctx, params)
				require.NoError(t, err)
				require.Equal(t, recurrent.ID, updated.ID)
				require.Equal(t, params.Description, updated.Description)
				require.Equal(t, params.Amount, updated.Amount)
				require.Equal(t, params.Period, updated.Period)
				require.True(t, updated.LastCopyCreatedAt.Valid)
				require.Equal(t, lastCopy.Int64, updated.LastCopyCreatedAt.Int64)
			},
		},
		{
			"should_fail_not_found",
			func(t *testing.T) {
				params := repo.UpdateRecurrentExpenseParams{
					ID:                -1,
					UserID:            user.ID,
					Description:       "Missing",
					Amount:            100,
					Period:            1,
					LastCopyCreatedAt: sql.NullInt64{Valid: false},
				}
				_, err := f.Store.UpdateRecurrentExpense(ctx, params)
				require.ErrorIs(t, err, logic.ErrNotFound)
			},
		},
		{
			"should_fail_validation",
			func(t *testing.T) {
				params := repo.UpdateRecurrentExpenseParams{
					ID:                recurrent.ID,
					UserID:            user.ID,
					Description:       "ab",
					Amount:            0,
					Period:            0,
					LastCopyCreatedAt: sql.NullInt64{Valid: false},
				}
				_, err := f.Store.UpdateRecurrentExpense(ctx, params)
				require.ErrorIs(t, err, logic.ErrValidationFailed)
				require.Contains(t, err.Error(), "[Description:min]")
				require.Contains(t, err.Error(), "[Amount:required]")
				require.Contains(t, err.Error(), "[Period:required]")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateLastCopyCreatedAt(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "updatelastcopyuser",
		Email:                "updatelastcopy@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Update Last Copy Category")
	recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Yearly subscription",
		Amount:      5000,
		Period:      12,
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_update_last_copy_created_at",
			func(t *testing.T) {
				copyDate := time.Now().Unix()
				updated, err := f.Store.UpdateLastCopyCreatedAt(ctx, recurrent, copyDate)
				require.NoError(t, err)
				require.True(t, updated.LastCopyCreatedAt.Valid)
				require.Equal(t, copyDate, updated.LastCopyCreatedAt.Int64)

				reloaded, err := f.Store.FindRecurrentExpense(ctx, recurrent.ID, user.ID)
				require.NoError(t, err)
				require.True(t, reloaded.LastCopyCreatedAt.Valid)
				require.Equal(t, copyDate, reloaded.LastCopyCreatedAt.Int64)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestCreateExpenseFromPeriod(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	user := f.User(t, logic.SignUpParams{
		Username:             "createfromperioduser",
		Email:                "createfromperiod@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Create From Period Category")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_create_expense_when_no_last_copy",
			func(t *testing.T) {
				recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
					CategoryID:  category.ID,
					Description: "Monthly rent",
					Amount:      100000,
					Period:      1,
				})

				expense, err := f.Store.CreateExpenseFromPeriod(ctx, recurrent)
				require.NoError(t, err)
				require.Equal(t, recurrent.Description, expense.Description)
				require.Equal(t, recurrent.Amount, expense.Amount)
				require.Equal(t, user.ID, expense.UserID)
				require.Equal(t, category.ID, expense.CategoryID)

				reloaded, err := f.Store.FindRecurrentExpense(ctx, recurrent.ID, user.ID)
				require.NoError(t, err)
				require.True(t, reloaded.LastCopyCreatedAt.Valid)
			},
		},
		{
			"should_fail_when_period_over_24",
			func(t *testing.T) {
				recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
					CategoryID:  category.ID,
					Description: "Annual fee",
					Amount:      2500,
					Period:      12,
				})

				_, err := f.Store.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
					ID:                recurrent.ID,
					UserID:            user.ID,
					Description:       recurrent.Description,
					Amount:            recurrent.Amount,
					Period:            25,
					LastCopyCreatedAt: sql.NullInt64{Valid: true, Int64: time.Now().AddDate(0, -30, 0).Unix()},
				})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			"should_fail_when_within_period",
			func(t *testing.T) {
				recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
					CategoryID:  category.ID,
					Description: "Insurance",
					Amount:      8000,
					Period:      3,
				})

				lastCopy := time.Now().AddDate(0, -2, 0).Unix()
				updated, err := f.Store.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
					ID:                recurrent.ID,
					UserID:            user.ID,
					Description:       recurrent.Description,
					Amount:            recurrent.Amount,
					Period:            recurrent.Period,
					LastCopyCreatedAt: sql.NullInt64{Valid: true, Int64: lastCopy},
				})
				require.NoError(t, err)

				_, err = f.Store.CreateExpenseFromPeriod(ctx, updated)
				require.ErrorIs(t, err, logic.ErrRecordAlreadyExist)
			},
		},
		{
			"should_create_expense_when_period_passed",
			func(t *testing.T) {
				recurrent := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
					CategoryID:  category.ID,
					Description: "License",
					Amount:      3000,
					Period:      2,
				})

				lastCopy := time.Now().AddDate(0, -3, 0).Unix()
				updated, err := f.Store.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
					ID:                recurrent.ID,
					UserID:            user.ID,
					Description:       recurrent.Description,
					Amount:            recurrent.Amount,
					Period:            recurrent.Period,
					LastCopyCreatedAt: sql.NullInt64{Valid: true, Int64: lastCopy},
				})
				require.NoError(t, err)

				expense, err := f.Store.CreateExpenseFromPeriod(ctx, updated)
				require.NoError(t, err)
				require.Equal(t, recurrent.Description, expense.Description)
				require.Equal(t, recurrent.Amount, expense.Amount)
				require.Equal(t, user.ID, expense.UserID)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
