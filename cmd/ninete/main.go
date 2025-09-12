// Package main
package main

import (
	"database/sql"
	"os"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
)

func main() {
	var exitCode int

	exitCode, err := start()
	if err != nil {
		prog.NewLogger().Error("%v", err)
	}

	os.Exit(exitCode)
}

func start() (int, error) {
	app, err := prog.Load()
	if err != nil {
		return 1, err
	}

	app.Logger.Log("Booting up application...")

	sqlDB, err := db.Open()
	if err != nil {
		return 1, err
	}
	defer closeDB(sqlDB)

	queries := repo.New(sqlDB)

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

func closeDB(sqlDB *sql.DB) {
	if err := sqlDB.Close(); err != nil {
		prog.NewLogger().Log("failed to close database: %v", err)
	}
}
