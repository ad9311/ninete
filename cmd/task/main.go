// Package main for running maintenance tasks
package main

import (
	"context"
	"os"
	"time"

	"github.com/ad9311/ninete/internal/cmd"
	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/task"
)

func main() {
	app, err := prog.Load()
	if err != nil {
		prog.QuickLogger().Errorf("%v", err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		app.Logger.Error(err)
	}

	queries := repo.New(app, sqlDB)

	store, err := logic.New(app, queries)
	if err != nil {
		app.Logger.Error(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	tc := task.Config{
		App:     app,
		SQLDB:   sqlDB,
		Store:   store,
		Context: ctx,
	}

	code, err := cmd.Run(os.Args[0], []*cmd.Command{
		{
			Name:        "test",
			Description: "Run test code",
			Run: func(_ []string) error {
				return tc.RunTestCode()
			},
		},
		{
			Name:        "create_categories",
			Description: "Create categories",
			Run: func(_ []string) error {
				return tc.CreateCategories()
			},
		},
	})
	if err != nil {
		app.Logger.Error(err)
	}

	cancel()
	os.Exit(code)
}
