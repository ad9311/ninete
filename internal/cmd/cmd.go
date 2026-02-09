package cmd

import (
	"fmt"
	"os"
	"sort"
)

type Command struct {
	Name        string
	Description string
	Run         func() error
}

type Registry struct {
	cmds  map[string]*Command
	usage string
}

func Run(appName string, cmds []*Command) (int, error) {
	reg := NewRegistry().WithUsageName(appName)

	cmds = append(cmds, &Command{
		Name:        "help",
		Description: "Prints all available commands",
		Run: func() error {
			reg.printUsage()

			return nil
		},
	})

	for _, c := range cmds {
		if err := reg.Register(c); err != nil {
			return 2, err
		}
	}

	return reg.Execute(os.Args[1:])
}

func (r *Registry) WithUsageName(name string) *Registry {
	r.usage = name

	return r
}

func NewRegistry() *Registry {
	return &Registry{
		cmds: make(map[string]*Command),
	}
}

func (r *Registry) Register(cmd *Command) error {
	if command, exists := r.cmds[cmd.Name]; exists {
		return fmt.Errorf("%w, '%s'", ErrCommandExists, command.Name)
	}
	r.cmds[cmd.Name] = cmd

	return nil
}

func (r *Registry) Execute(args []string) (int, error) {
	if len(args) == 0 {
		r.printUsage()

		return 1, nil
	}

	name := args[0]
	cmd, ok := r.cmds[name]
	if !ok {
		r.printUsage()

		return 1, fmt.Errorf("%w, '%s'", ErrUnknownCommand, name)
	}

	if err := cmd.Run(); err != nil {
		msg := "command " + "'" + name + "'" + " failed,"

		return 2, fmt.Errorf("%s %w", msg, err)
	}

	return 0, nil
}

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
