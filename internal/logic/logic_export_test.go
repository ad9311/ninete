package logic_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestExportExpenses(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "export_user_1",
		Email:        "export_user_1@example.com",
		PasswordHash: []byte("export_user_hash_1"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "export_user_2",
		Email:        "export_user_2@example.com",
		PasswordHash: []byte("export_user_hash_2"),
	})
	category := s.CreateCategory(t, "export_category_1")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_return_only_user_expenses",
			fn: func(t *testing.T) {
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "mine_1", 100, 1735689600, nil))
				s.CreateExpense(t, otherUser.ID, newExpenseParams(category.ID, "theirs_1", 200, 1735689600, nil))

				out, err := s.Store.ExportExpenses(ctx, user.ID)
				require.NoError(t, err)

				for _, e := range out {
					require.NotEqual(t, "theirs_1", e.Description)
				}
				descs := make([]string, 0, len(out))
				for _, e := range out {
					descs = append(descs, e.Description)
				}
				require.Contains(t, descs, "mine_1")
			},
		},
		{
			name: "should_attach_category_to_each_expense",
			fn: func(t *testing.T) {
				s.CreateExpense(t, user.ID, newExpenseParams(category.ID, "with_cat", 100, 1735689600, nil))

				out, err := s.Store.ExportExpenses(ctx, user.ID)
				require.NoError(t, err)
				require.NotEmpty(t, out)

				for _, e := range out {
					if e.Description == "with_cat" {
						require.NotNil(t, e.Category)
						require.Equal(t, "export_category_1", e.Category.Name)

						return
					}
				}
				t.Fatalf("expense 'with_cat' not found in export")
			},
		},
		{
			name: "should_attach_sorted_tags",
			fn: func(t *testing.T) {
				s.CreateExpense(
					t, user.ID,
					newExpenseParams(category.ID, "tagged", 100, 1735689600, []string{"zebra", "alpha", "mango"}),
				)

				out, err := s.Store.ExportExpenses(ctx, user.ID)
				require.NoError(t, err)

				for _, e := range out {
					if e.Description == "tagged" {
						require.Equal(t, []string{"alpha", "mango", "zebra"}, e.Tags)

						return
					}
				}
				t.Fatalf("expense 'tagged' not found in export")
			},
		},
		{
			name: "should_return_empty_slice_when_user_has_no_expenses",
			fn: func(t *testing.T) {
				lonelyUser := s.CreateUser(t, repo.InsertUserParams{
					Username:     "export_user_3",
					Email:        "export_user_3@example.com",
					PasswordHash: []byte("export_user_hash_3"),
				})

				out, err := s.Store.ExportExpenses(ctx, lonelyUser.ID)
				require.NoError(t, err)
				require.Empty(t, out)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
