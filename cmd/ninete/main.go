package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
)

func main() {
	var exitCode int

	// Handled before prog.Load, which requires ENV to be set: asking an installed
	// binary what it is must work without the service's environment.
	//
	// This is the second of two implementations. `migrate` and `task` get their
	// `version` command from the registry in internal/cmd, which this binary does
	// not use — it is a server, not a subcommand dispatcher, so a registry built
	// for one command would buy nothing. Keep the output identical to
	// cmd.Run's: both print prog.VersionString() and nothing else.
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(prog.VersionString())

		os.Exit(0)
	}

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
	app.Logger.Logf("Booting up application... %s", prog.VersionString())

	sqlDB, err := db.Open()
	if err != nil {
		return 1, err
	}
	defer func() {
		if err := db.Optimize(context.Background(), sqlDB); err != nil {
			app.Logger.Errorf("%v", err)
		}

		if err := sqlDB.Close(); err != nil {
			app.Logger.Errorf("failed to close database: %v", err)
		}
	}()

	queries := repo.New(app, sqlDB)

	store := logic.New(app, queries)

	server := serve.New(app, store, sqlDB)

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
