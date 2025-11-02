package logic_test

import (
	"os"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if code := testhelper.SetUpPackageTest("logic_test"); code > 0 {
		os.Exit(code)
	}

	os.Exit(m.Run())
}

func newTestStore(t *testing.T) *logic.Store {
	t.Helper()

	app, err := prog.Load()
	require.NoError(t, err)

	sqlDB, err := db.Open()
	require.NoError(t, err)

	queries := repo.New(app, sqlDB)

	store, err := logic.New(app, queries)
	require.NoError(t, err)

	return store
}
