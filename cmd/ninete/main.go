// Package main
package main

import (
	"os"

	"github.com/ad9311/ninete/internal/app"
	"github.com/ad9311/ninete/internal/csl"
	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
	"github.com/ad9311/ninete/internal/srv"
)

func main() {
	var exitCode int

	exitCode, err := load()
	if err != nil {
		csl.NewError("falied to boot application, %v", err)
	}

	os.Exit(exitCode)
}

func load() (int, error) {
	_, err := app.Load()
	if err != nil {
		return 1, err
	}

	sqlDB, err := db.Open()
	if err != nil {
		return 1, err
	}
	defer sqlDB.Close()

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
