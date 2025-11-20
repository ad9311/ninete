// Package main for running maintenance tasks
package main

import (
	"os"

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
		prog.NewLogger(prog.LogOptions{
			EnableColor: true,
		}).Errorf("%v", err)
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

	code, err := cmd.Run(os.Args[0], []*cmd.Command{
		{
			Name:        "dev",
			Description: "Run test code",
			Run: func(_ []string) error {
				return task.RunDev(store)
			},
		},
	})
	if err != nil {
		app.Logger.Error(err)
	}

	os.Exit(code)
}
