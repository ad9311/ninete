package main

import (
	"os"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
)

func main() {
	var exitCode int

	app, err := prog.Load()
	if err != nil {
		prog.QuickLogger().Errorf("failed to load app configuration: %v", err)

		os.Exit(1)
	}

	exitCode, err = start(app)
	if err != nil {
		app.Logger.Errorf("%v", err)
	}

	os.Exit(exitCode)
}

func start(app *prog.App) (int, error) {
	app.Logger.Log("Booting up application...")

	sqlDB, err := db.Open()
	if err != nil {
		return 1, err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			app.Logger.Errorf("failed to close database: %v", err)
		}
	}()

	queries := repo.New(app, sqlDB)

	store := logic.New(app, queries)

	server, err := serve.New(app, store)
	if err != nil {
		return 1, err
	}

	err = server.LoadTemplates()
	if err != nil {
		return 1, err
	}

	err = server.Start()
	if err != nil {
		return 1, err
	}

	return 0, nil
}
