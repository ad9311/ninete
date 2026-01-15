package task_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/seed"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if code := testhelper.SetUpPackageTest("task_test"); code > 0 {
		os.Exit(code)
	}

	os.Exit(m.Run())
}

func TestRunTestCode(t *testing.T) {
	f := testhelper.NewFactory(t)
	err := f.TaskConfig.RunTestCode()
	require.NoError(t, err)
}

func TestCreateCategories(t *testing.T) {
	f := testhelper.NewFactory(t)

	err := f.TaskConfig.CreateCategories()
	require.NoError(t, err)

	categories, err := f.Store.FindCategories(t.Context())
	require.NoError(t, err)
	categoryNames := seed.CategoryNames()
	size := len(categoryNames)
	require.Len(t, categories, size)
}

func TestCreateExpensesFromRecurrent(t *testing.T) {
	f := testhelper.NewFactory(t)
	ctx := t.Context()

	user := f.User(t, logic.SignUpParams{
		Username:             "taskrecurrentuser",
		Email:                "taskrecurrent@email.com",
		Password:             "123456789",
		PasswordConfirmation: "123456789",
	})
	category := f.Category(t, "Task Recurrent Category")
	recurrentDue := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Task internet",
		Amount:      5500,
		Period:      1,
	})
	_, err := f.Store.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
		ID:                recurrentDue.ID,
		UserID:            user.ID,
		Description:       recurrentDue.Description,
		Amount:            recurrentDue.Amount,
		Period:            recurrentDue.Period,
		LastCopyCreatedAt: sql.NullInt64{Valid: true, Int64: time.Now().AddDate(0, -1, 0).Unix()},
	})
	require.NoError(t, err)
	recurrentDueSecond := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Task mobile",
		Amount:      4200,
		Period:      1,
	})
	_, err = f.Store.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
		ID:                recurrentDueSecond.ID,
		UserID:            user.ID,
		Description:       recurrentDueSecond.Description,
		Amount:            recurrentDueSecond.Amount,
		Period:            recurrentDueSecond.Period,
		LastCopyCreatedAt: sql.NullInt64{Valid: true, Int64: time.Now().AddDate(0, -1, 0).Unix()},
	})
	require.NoError(t, err)
	recurrentNotDue := f.RecurrentExpense(t, user.ID, logic.RecurrentExpenseParams{
		CategoryID:  category.ID,
		Description: "Not due yet",
		Amount:      2500,
		Period:      2,
	})
	_, err = f.Store.UpdateRecurrentExpense(ctx, repo.UpdateRecurrentExpenseParams{
		ID:                recurrentNotDue.ID,
		UserID:            user.ID,
		Description:       recurrentNotDue.Description,
		Amount:            recurrentNotDue.Amount,
		Period:            recurrentNotDue.Period,
		LastCopyCreatedAt: sql.NullInt64{Valid: true, Int64: time.Now().Unix()},
	})
	require.NoError(t, err)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_create_due_expenses_and_skip_not_due",
			func(t *testing.T) {
				err := f.TaskConfig.CreateExpensesFromRecurrent(ctx)
				require.NoError(t, err)

				expenses, err := f.Store.FindExpenses(ctx, repo.QueryOptions{
					Filters: repo.Filters{
						FilterFields: []repo.FilterField{
							{Name: "user_id", Value: user.ID, Operator: "="},
						},
					},
				})
				require.NoError(t, err)
				require.Len(t, expenses, 2)
				descriptions := []string{expenses[0].Description, expenses[1].Description}
				require.ElementsMatch(t, []string{
					recurrentDue.Description,
					recurrentDueSecond.Description,
				}, descriptions)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteExpiredRefreshTokens(t *testing.T) {
	f := testhelper.NewFactory(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_delete_expired_tokens",
			func(t *testing.T) {
				ctx := t.Context()

				user := f.User(t, logic.SignUpParams{
					Username:             "taskdeleteexpired",
					Email:                "taskdeleteexpired@email.com",
					Password:             "123456789",
					PasswordConfirmation: "123456789",
				})
				expiredToken := f.RefreshToken(t, user.ID)
				validToken := f.RefreshToken(t, user.ID)
				f.SetRefreshTokenExpiry(t, expiredToken.Value, time.Now().Add(-time.Hour).Unix())

				err := f.TaskConfig.DeleteExpiredRefreshTokens(ctx)
				require.NoError(t, err)

				_, err = f.Store.FindRefreshToken(ctx, expiredToken.Value)
				require.ErrorIs(t, err, logic.ErrNotFound)

				_, err = f.Store.FindRefreshToken(ctx, validToken.Value)
				require.NoError(t, err)
			},
		},
		{
			"should_do_nothing_when_none_expired",
			func(t *testing.T) {
				ctx := t.Context()

				user := f.User(t, logic.SignUpParams{
					Username:             "tasknoexpired",
					Email:                "tasknoexpired@email.com",
					Password:             "123456789",
					PasswordConfirmation: "123456789",
				})
				_ = f.RefreshToken(t, user.ID)

				err := f.TaskConfig.DeleteExpiredRefreshTokens(ctx)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
