package logic_test

import (
	"fmt"
	"testing"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

// TestPaginationIsStableAcrossPages walks every page of a listing whose sort
// column is identical on all rows, and asserts the pages partition the rows.
//
// This is an invariant guard, not a reproduction: SQLite happens to break the
// ties consistently for the plan this query gets today, so it passes without the
// "id" tiebreaker too. TestSortingBuild in internal/repo is the test that fails
// when the tiebreaker goes missing.
func TestPaginationIsStableAcrossPages(t *testing.T) {
	const (
		expenseCount = 40
		perPage      = 7
		sameDate     = int64(1735689600)
	)

	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "pagination_user",
		Email:        "pagination_user@example.com",
		PasswordHash: []byte("pagination_hash"),
	})
	category := s.CreateCategory(t, "pagination_category")

	for i := range expenseCount {
		description := fmt.Sprintf("paged_%04d", i)
		s.CreateExpense(t, user.ID, newExpenseParams(category.ID, description, 100, sameDate, nil))
	}

	seen := make(map[int]int, expenseCount)

	for page := 1; page <= (expenseCount+perPage-1)/perPage; page++ {
		opts := repo.QueryOptions{
			Sorting:    repo.Sorting{Field: "date", Order: "DESC"},
			Pagination: repo.Pagination{Page: page, PerPage: perPage},
			Filters: repo.Filters{
				FilterFields: []repo.FilterField{
					{Name: "user_id", Value: user.ID, Operator: "="},
				},
				Connector: "AND",
			},
		}

		expenses, err := s.Store.FindExpenses(ctx, opts)
		require.NoError(t, err)

		for _, e := range expenses {
			seen[e.ID]++
		}
	}

	require.Len(t, seen, expenseCount, "pages did not cover every expense exactly once")

	for id, count := range seen {
		require.Equal(t, 1, count, "expense %d appeared on more than one page", id)
	}
}
