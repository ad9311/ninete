// Package main is the entry point for the CLI application.
package main

import (
	"log"
	"os"

	"github.com/ad9311/go-api-base/cmd"
	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/db"
	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/server"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/ad9311/go-api-base/internal/task"
)

// main initializes the command registry, registers available commands, and executes the selected command.
func main() {
	r := cmd.NewRegistry().WithUsageName(os.Args[0])

	commands := []func() *cmd.Command{
		serveCommand,
		runUpCommand,
		runDownCommand,
		runTaskCommand,
	}

	for _, c := range commands {
		if err := r.Register(c()); err != nil {
			log.Println(err)

			os.Exit(2)
		}
	}

	exitCode, err := r.Execute(os.Args[1:])
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitCode)
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
