package logic_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestExpenseBudgets(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	user := s.CreateAuthUser(t, "budget_logic_user", "budget_logic_user@example.com", "budget_password_1")
	other := s.CreateAuthUser(t, "budget_other_user", "budget_other_user@example.com", "budget_password_2")
	category := s.CreateCategory(t, "budget logic category")
	otherCategory := s.CreateCategory(t, "budget logic category two")

	budgetFor := func(t *testing.T, userID, categoryID int) (repo.ExpenseBudget, bool) {
		t.Helper()

		budgets, err := s.Store.FindExpenseBudgets(ctx, userID)
		require.NoError(t, err)

		for _, b := range budgets {
			if b.CategoryID == categoryID {
				return b, true
			}
		}

		return repo.ExpenseBudget{}, false
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_insert_then_update_the_same_category",
			fn: func(t *testing.T) {
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{category.ID: 50000})

				budget, ok := budgetFor(t, user.ID, category.ID)
				require.True(t, ok)
				require.Equal(t, uint64(50000), budget.Amount)

				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{category.ID: 60000})

				updated, ok := budgetFor(t, user.ID, category.ID)
				require.True(t, ok)
				require.Equal(t, uint64(60000), updated.Amount)
				require.Equal(t, budget.ID, updated.ID)
			},
		},
		{
			name: "should_delete_the_budget_when_the_amount_is_cleared",
			fn: func(t *testing.T) {
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{otherCategory.ID: 25000})
				_, ok := budgetFor(t, user.ID, otherCategory.ID)
				require.True(t, ok)

				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{otherCategory.ID: 0})

				_, ok = budgetFor(t, user.ID, otherCategory.ID)
				require.False(t, ok)
			},
		},
		{
			name: "should_scope_budgets_to_their_owner",
			fn: func(t *testing.T) {
				s.SaveExpenseBudgets(t, other.ID, map[int]uint64{category.ID: 11100})

				ownerBudgets, err := s.Store.FindExpenseBudgets(ctx, user.ID)
				require.NoError(t, err)
				for _, b := range ownerBudgets {
					require.Equal(t, user.ID, b.UserID)
					require.NotEqual(t, uint64(11100), b.Amount)
				}

				otherBudget, ok := budgetFor(t, other.ID, category.ID)
				require.True(t, ok)
				require.Equal(t, uint64(11100), otherBudget.Amount)
			},
		},
		{
			name: "should_delete_every_budget_for_one_user_only",
			fn: func(t *testing.T) {
				s.SaveExpenseBudgets(t, user.ID, map[int]uint64{category.ID: 70000})
				s.SaveExpenseBudgets(t, other.ID, map[int]uint64{category.ID: 80000})

				require.NoError(t, s.Store.DeleteAllExpenseBudgets(ctx, user.ID))

				ownerBudgets, err := s.Store.FindExpenseBudgets(ctx, user.ID)
				require.NoError(t, err)
				require.Empty(t, ownerBudgets)

				otherBudgets, err := s.Store.FindExpenseBudgets(ctx, other.ID)
				require.NoError(t, err)
				require.Len(t, otherBudgets, 1)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
