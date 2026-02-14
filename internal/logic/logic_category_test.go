package logic_test

import (
	"testing"

	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateCategory(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_category",
			fn: func(t *testing.T) {
				category, err := s.Store.CreateCategory(ctx, "new category 1")
				require.NoError(t, err)
				require.Positive(t, category.ID)
				require.Equal(t, "new category 1", category.Name)
				require.Equal(t, "newCategory1", category.UID)
				require.NotZero(t, category.CreatedAt)
				require.NotZero(t, category.UpdatedAt)
			},
		},
		{
			name: "should_fail_with_duplicate_name",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateCategory(ctx, "new category 2")
				require.NoError(t, err)

				_, err = s.Store.CreateCategory(ctx, "new category 2")
				require.Error(t, err)
			},
		},
		{
			name: "should_fail_with_duplicate_uid",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateCategory(ctx, "new-category-3")
				require.NoError(t, err)

				_, err = s.Store.CreateCategory(ctx, "new category 3")
				require.Error(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindCategories(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_created_categories",
			fn: func(t *testing.T) {
				categoryOne := s.CreateCategory(t, "new category 4")
				categoryTwo := s.CreateCategory(t, "new category 5")

				categories, err := s.Store.FindCategories(ctx)
				require.NoError(t, err)
				require.NotEmpty(t, categories)

				var foundOne bool
				var foundTwo bool

				for _, category := range categories {
					if category.ID == categoryOne.ID {
						foundOne = true
					}
					if category.ID == categoryTwo.ID {
						foundTwo = true
					}
				}

				require.True(t, foundOne)
				require.True(t, foundTwo)
			},
		},
		{
			name: "should_return_categories_sorted_by_name",
			fn: func(t *testing.T) {
				aCategory := s.CreateCategory(t, "new category 6")
				bCategory := s.CreateCategory(t, "new category 7")

				categories, err := s.Store.FindCategories(ctx)
				require.NoError(t, err)

				indexesByID := map[int]int{}
				for i, category := range categories {
					indexesByID[category.ID] = i
				}

				require.Less(t, indexesByID[aCategory.ID], indexesByID[bCategory.ID])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
