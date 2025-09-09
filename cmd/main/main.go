// Package main is the entry point for the CLI application.
package main

import (
	"os"

	"github.com/ad9311/go-api-base/cmd"
	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/console"
	"github.com/ad9311/go-api-base/internal/db"
	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/server"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/ad9311/go-api-base/internal/task"
)

// main initializes the command registry, registers available commands, and executes the selected command.
func main() {
	logger := console.New(nil, nil, false)

	r := cmd.NewRegistry().WithUsageName(os.Args[0])

	commands := []func() *cmd.Command{
		testDevCommand,
		serveCommand,
		runUpCommand,
		runDownCommand,
		runStatusCommand,
		runTaskCommand,
	}

	for _, c := range commands {
		if err := r.Register(c()); err != nil {
			logger.Error("%v", err)

			os.Exit(2)
		}
	}

	exitCode, err := r.Execute(os.Args[1:])
	if err != nil {
		logger.Error("%v", err)
	}

	os.Exit(exitCode)
}

// testDevCommand returns a CLI command intended for development and testing purposes only.
// Use this command to quickly run or debug code snippets during development. Not for production use.
func testDevCommand() *cmd.Command {
	return &cmd.Command{
		Name:        "dev_func",
		Description: "Run a development/test function (for developer use only)",
		Run: func(_ []string) error {
			// Place any code here you want to test during development.
			return nil
		},
	}
}

// serveCommand returns the CLI command for running the server.
func serveCommand() *cmd.Command {
	return &cmd.Command{
		Name:        "server",
		Description: "Run the server",
		Run:         serve,
	}
}

// runUpCommand returns the CLI command for applying all up migrations.
func runUpCommand() *cmd.Command {
	return &cmd.Command{
		Name:        "migrate",
		Description: "Apply all up migrations",
		Run:         db.RunMigrationsUp,
	}
}

// runDownCommand returns the CLI command for applying one migration down.
func runDownCommand() *cmd.Command {
	return &cmd.Command{
		Name:        "migrate-down",
		Description: "Apply one migration down",
		Run:         db.RunMigrationsDown,
	}
}

// runStatusCommand returns the CLI command for printing the migrations status.
func runStatusCommand() *cmd.Command {
	return &cmd.Command{
		Name:        "status",
		Description: "Print database status",
		Run:         db.PrintStatus,
	}
}

// runTaskCommand returns the CLI command for running a maintenance or utility task.
func runTaskCommand() *cmd.Command {
	return &cmd.Command{
		Name:        "task",
		Description: "Run a task",
		Run:         task.RunTask,
	}
}

// serve starts the HTTP server using the loaded configuration and database connection.
func serve(_ []string) error {
	config, err := app.LoadConfig()
	if err != nil {
		return err
	}

	pool, err := db.Connect(config)
	if err != nil {
		return err
	}
	defer pool.Close()

	store, err := service.New(config, pool)
	if err != nil {
		return err
	}

	server := server.New(config, store)
	if err := server.Start(); err != nil {
		return errs.WrapErrorWithMessage("failed to start server", err)
	}

	return nil
}
