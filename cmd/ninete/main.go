// Package main
package main

import (
	"database/sql"
	"os"

	"github.com/ad9311/ninete/internal/app"
	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
	"github.com/ad9311/ninete/internal/srv"
)

func main() {
	var exitCode int

	app.Log("Booting up application...")

	exitCode, err := start()
	if err != nil {
		app.LogError("%v", err)
	}

	os.Exit(exitCode)
}

func start() (int, error) {
	if err := app.Load(); err != nil {
		return 1, err
	}

	sqlDB, err := db.Open()
	if err != nil {
		return 1, err
	}
	defer closeDB(sqlDB)

	queries := repo.New(sqlDB)

	store, err := srv.New(queries)
	if err != nil {
		return 1, err
	}

	server, err := serve.New(store)
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
		app.Log("failed to close database: %v", err)
	}
}
