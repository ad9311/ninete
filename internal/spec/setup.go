package spec

import (
	"os"
	"path/filepath"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/prog"
)

func SetupPackageTest(dbName string) int {
	app, err := prog.Load()
	if err != nil {
		prog.QuickLogger().Errorf("failed to load app configuration: %v", err)

		return 1
	}

	root, ok := prog.FindRoot()
	if !ok {
		app.Logger.Errorf("failed to failed root directory: %v", os.ErrNotExist)

		return 1
	}

	testDBDir := filepath.Join(root, "data", "db", "test")
	if err := os.MkdirAll(testDBDir, 0o750); err != nil {
		app.Logger.Errorf("failed to create test database directory: %v", err)

		return 1
	}

	testDBURL := filepath.Join(testDBDir, dbName)

	if err := os.Remove(testDBURL); err != nil && !os.IsNotExist(err) {
		app.Logger.Errorf("failed to reset package test database: %v", err)

		return 1
	}

	if err := os.Setenv("DATABASE_URL", testDBURL); err != nil {
		app.Logger.Errorf("failed to set DATABASE_URL: %v", err)

		return 1
	}

	if err := db.RunMigrationsUp(); err != nil {
		app.Logger.Errorf("failed to run package test migrations: %v", err)

		return 1
	}

	return 0
}
