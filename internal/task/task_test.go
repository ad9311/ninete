package task_test

import (
	"os"
	"testing"

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
