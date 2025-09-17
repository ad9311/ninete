// Package main
package main

import (
	"log"
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
		log.Fatalf("failed to load app configuration: %v", err)
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

	store, err := logic.New(app, queries)
	if err != nil {
		return 1, err
	}

	server, err := serve.New(app, store)
	if err != nil {
		return 1, err
	}

	err = server.Start()
	if err != nil {
		return 1, err
	}

	return 0, nil
}
