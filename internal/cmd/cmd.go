// Package cmd provides a simple command registry and execution framework for building
// command-line interfaces (CLIs). It allows registration of commands with metadata and
// execution logic, handles command dispatch based on user input, and prints usage/help
// information.
package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/ad9311/ninete/internal/errs"
)

// Command represents a CLI command's metadata and execution logic.
type Command struct {
	Name        string
	Description string
	Run         func(args []string) error
}

// Registry holds registered commands and an optional usage name for help output.
type Registry struct {
	cmds  map[string]*Command
	usage string
}

// Run initializes a new command registry with the provided application name,
// registers the given list of commands, and executes the registry with the
// command-line arguments (excluding the program name). It returns an exit code
// and an error, if any occurred during registration or execution.
func Run(appName string, cmds []*Command) (int, error) {
	reg := NewRegistry().WithUsageName(appName)

	for _, c := range cmds {
		if err := reg.Register(c); err != nil {
			return 2, err
		}
	}

	return reg.Execute(os.Args[1:])
}

// WithUsageName sets the binary name shown in usage/help output.
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
func (r *Registry) Register(cmd *Command) error {
	if command, exists := r.cmds[cmd.Name]; exists {
		return fmt.Errorf("%w: %s", errs.ErrCommandExists, command.Name)
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

		return 1, fmt.Errorf("%w: %s", errs.ErrUnknownCommand, name)
	}

	if err := cmd.Run(args[1:]); err != nil {
		msg := "command " + name + " failed with: "

		return 2, fmt.Errorf("%s %w", msg, err)
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
