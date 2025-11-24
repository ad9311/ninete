package logic_test

import (
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/testhelper"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestCreateCategory(t *testing.T) {
	ctx := t.Context()
	f := testhelper.NewFactory(t)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_create_category",
			func(t *testing.T) {
				categoryName := "House Bills"
				category := f.Category(t, categoryName)

				expectedUID := strings.ReplaceAll(strings.ToLower(categoryName), " ", "_")

				require.Equal(t, categoryName, category.Name)
				require.Equal(t, expectedUID, category.UID)
				require.Positive(t, category.ID)
			},
		},
		{
			"should_fail_unique_name",
			func(t *testing.T) {
				categoryName := "Unique Name"
				uidOne := "unique_name_uid1"
				uidTwo := "unique_name_uid2"

				_, err := f.Store.CreateCategory(ctx, categoryName, uidOne)
				require.NoError(t, err)

				_, err = f.Store.CreateCategory(ctx, categoryName, uidTwo)
				requireUniqueConstraint(t, err, "categories.name")
			},
		},
		{
			"should_fail_unique_uid",
			func(t *testing.T) {
				nameOne := "Name One"
				nameTwo := "Name Two"
				uid := "same_uid"

				_, err := f.Store.CreateCategory(ctx, nameOne, uid)
				require.NoError(t, err)

				_, err = f.Store.CreateCategory(ctx, nameTwo, uid)
				requireUniqueConstraint(t, err, "categories.uid")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func requireUniqueConstraint(t *testing.T, err error, field string) {
	t.Helper()

	require.Error(t, err)

	var sqlErr sqlite3.Error
	require.ErrorAs(t, err, &sqlErr)
	require.Equal(t, sqlite3.ErrConstraint, sqlErr.Code)
	require.Equal(t, sqlite3.ErrConstraintUnique, sqlErr.ExtendedCode)
	require.Contains(t, err.Error(), field)
}
