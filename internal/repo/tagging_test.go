package repo_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestTagNamesByTargetID(t *testing.T) {
	cases := []struct {
		name string
		in   []repo.TagRow
		want map[int][]string
	}{
		{
			name: "should_return_empty_map_for_no_rows",
			in:   nil,
			want: map[int][]string{},
		},
		{
			name: "should_group_rows_by_target_id",
			in: []repo.TagRow{
				{TargetID: 1, TagName: "alpha"},
				{TargetID: 2, TagName: "beta"},
				{TargetID: 1, TagName: "gamma"},
			},
			want: map[int][]string{
				1: {"alpha", "gamma"},
				2: {"beta"},
			},
		},
		{
			name: "should_sort_tag_names_within_each_group",
			in: []repo.TagRow{
				{TargetID: 7, TagName: "zebra"},
				{TargetID: 7, TagName: "alpha"},
				{TargetID: 7, TagName: "mango"},
			},
			want: map[int][]string{
				7: {"alpha", "mango", "zebra"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := repo.TagNamesByTargetID(tc.in)
			require.Equal(t, tc.want, got)
		})
	}
}

// sqliteMaxVariableNumber is the parameter ceiling SQLite enforces per
// statement, as go-sqlite3 builds it. A statement carrying more is rejected
// outright with "too many SQL variables".
const sqliteMaxVariableNumber = 32766

// TestSelectTagRowsExceedsParameterLimit is a genuine reproduction: it passes
// more ids than SQLite accepts parameters in one statement. Without the
// chunking in SelectTagRows the whole slice lands in a single IN (...) list and
// the query fails.
//
// The ids do not need to exist — the limit is on the statement, not the data —
// so this stays fast where building 32767 real expenses would not. One real
// tagged expense rides along at the end of the slice to prove the batched
// version still returns its rows rather than swallowing them.
func TestSelectTagRowsExceedsParameterLimit(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "tag_rows_limit_user",
		Email:        "tag_rows_limit_user@example.com",
		PasswordHash: []byte("tag_rows_limit_hash"),
	})
	category := s.CreateCategory(t, "tag_rows_limit_category")

	expense := s.CreateExpense(t, user.ID, logic.ExpenseParams{
		ExpenseBaseParams: logic.ExpenseBaseParams{
			CategoryID:  category.ID,
			Description: "tag_rows_limit_expense",
			Amount:      100,
		},
		Date: 1735689600,
		Tags: []string{"tag_rows_limit_tag"},
	})

	// Padding ids are negative so they cannot collide with a real row.
	targetIDs := make([]int, 0, sqliteMaxVariableNumber+2)
	for i := range sqliteMaxVariableNumber + 1 {
		targetIDs = append(targetIDs, -(i + 1))
	}

	targetIDs = append(targetIDs, expense.ID)

	rows, err := s.Queries.SelectTagRows(ctx, repo.TaggableTypeExpense, "expenses", targetIDs, user.ID)
	require.NoError(t, err)
	require.Equal(t, []repo.TagRow{{TargetID: expense.ID, TagName: "tag_rows_limit_tag"}}, rows)
}
