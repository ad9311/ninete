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

type TaskFunc func(*prog.App, *logic.Store) error

func main() {
	code, err := cmd.Run(os.Args[0], taskCommands())
	if err != nil {
		prog.QuickLogger().Errorf("%v", err)
	}

	os.Exit(code)
}

func taskCommands() []*cmd.Command {
	return []*cmd.Command{
		{
			Name:        "testdev",
			Description: "Runs development-only task hooks",
			Run:         runTask(task.TestDev),
		},
	}
}

func runTask(fn TaskFunc) func() error {
	return func() error {
		return execTask(fn)
	}
}

func execTask(fn TaskFunc) error {
	app, err := prog.Load()
	if err != nil {
		return err
	}

	sqlDB, err := db.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			app.Logger.Errorf("failed to close database: %v", err)
		}
	}()

	queries := repo.New(app, sqlDB)
	store := logic.New(app, queries)

	return fn(app, store)
}
