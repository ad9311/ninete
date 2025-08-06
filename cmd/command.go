// Package cmd provides a composable framework for building command-line interfaces (CLIs).
// It defines a Command type and a Registry that maps command names to their handlers.
package cmd

import (
	"fmt"
	"sort"

	"github.com/ad9311/go-api-base/internal/errs"
)

// Command represents a CLI command's metadata and execution logic.
// Future commands register themselves via init() without modifying this file.
type Command struct {
	Name        string                    // Unique command name (e.g. "migrate:up")
	Description string                    // One-line description for help text
	Run         func(args []string) error // Execution function, receives any extra args
}

// Registry holds registered commands and an optional usage name for help output.
type Registry struct {
	cmds  map[string]*Command
	usage string // optional override for the binary name shown in usage
}

// WithUsageName sets the binary name shown in usage/help output.
// WithUsageName sets the binary name shown in usage/help output and returns the registry for chaining.
func (r *Registry) WithUsageName(name string) *Registry {
	r.usage = name

	return r
}

// NewRegistry creates and returns an empty command registry.
func NewRegistry() *Registry {
	return &Registry{
		cmds: make(map[string]*Command),
	}
}

// Register adds a new command to the registry.
// Returns an error if a command with the same name already exists.
func (r *Registry) Register(cmd *Command) error {
	if command, exists := r.cmds[cmd.Name]; exists {
		return errs.WrapMessageWithError(errs.ErrCommandAlreadyRegistered, command.Name)
	}
	r.cmds[cmd.Name] = cmd

	return nil
}

// Execute runs the command matching args[0].
// Returns a process exit code: 0 on success, 1 on usage/unknown command, 2 if the command returned an error.
func (r *Registry) Execute(args []string) (int, error) {
	if len(args) == 0 {
		r.printUsage()

		return 1, nil
	}
	name := args[0]
	cmd, ok := r.cmds[name]
	if !ok {
		r.printUsage()

		return 1, errs.WrapMessageWithError(errs.ErrUnknowCommand, name)
	}
	if err := cmd.Run(args[1:]); err != nil {
		msg := "command " + name + " failed with: "

		return 2, errs.WrapErrorWithMessage(msg, err)
	}

	return 0, nil
}

// printUsage prints usage information and a list of available commands to stdout.
func (r *Registry) printUsage() {
	bin := r.usage
	if bin == "" {
		bin = "app"
	}
	fmt.Printf("Usage: %s <command> [options]\n\n", bin)
	fmt.Println("Available commands:")

	names := make([]string, 0, len(r.cmds))
	for name := range r.cmds {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("  %-20s %s\n", name, r.cmds[name].Description)
	}
}
