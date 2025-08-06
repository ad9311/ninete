// Package task provides functionality to run maintenance and utility tasks for the application.
package task

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/db"
	"github.com/ad9311/go-api-base/internal/service"
)

// funcTask represents a function that performs a task and returns an error if it fails.
type funcTask func() error

// task is used to run tasks with the provided configuration and service store.
type task struct {
	config       *app.Config    // Application configuration
	serviceStore *service.Store // Service layer store
	reader       io.Reader
}

// taskDetails contains the description and function for a specific task.
type taskDetails struct {
	description string   // Description of the task
	funcTask    funcTask // Function to execute the task
}

// Error variables for tasks
var (
	ErrUnexpectedArgsNumber = errors.New("unexpected number of arguments")
	ErrUnknownTaskCommand   = errors.New("unknown task command")
	ErrWrongEnvForTasks     = errors.New("tasks can only run in maintenance mode")
	ErrEmptyCSVFile         = errors.New("empty CSV file")
	ErrWrongNumOfColumns    = errors.New("wrong number of columns")
	ErrEmptyRoleName        = errors.New("role name cannot be empty")
)

// RunTask runs a task by its name, passing command-line arguments. Returns an error if the task fails or is not found.
func RunTask(args []string) error {
	task, err := setUp()
	if err != nil {
		return err
	}

	length := len(args)
	if length == 0 || length > 1 {
		return fmt.Errorf("%w, got %d, want 1", ErrUnexpectedArgsNumber, length)
	}

	name := args[0]
	_, details := task.taskCollection()
	td, ok := details[name]
	if !ok {
		return fmt.Errorf("%w %s", ErrUnknownTaskCommand, name)
	}

	if err := td.funcTask(); err != nil {
		return err
	}

	return nil
}

// printTasks prints all available tasks and their descriptions to the console.
func (t *task) printTasks() error {
	var task task
	names, details := task.taskCollection()

	fmt.Printf("Usage: task <name>\n\n")
	fmt.Println("Available tasks:")

	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("  %-30s %s\n", name, details[name].description)
	}

	return nil
}

// taskCollection returns a sorted list of task names and a map of their details.
func (t *task) taskCollection() ([]string, map[string]taskDetails) {
	mapDetails := map[string]taskDetails{
		"print": {
			description: "Prints all tasks",
			funcTask:    t.printTasks,
		},
		"ping_db": {
			description: "Pings the database",
			funcTask:    t.pingDatabaseTask,
		},
		"delete_expired_tokens": {
			description: "Deletes all expired refresh tokens",
			funcTask:    t.deleteExpiredTokensTask,
		},
		"create_admin_role": {
			description: "Creates the admin role",
			funcTask:    t.createAdminRoleTask,
		},
		"add_roles_to_users": {
			description: "Adds roles to users",
			funcTask:    t.addRolesToUsersTask,
		},
		"create_role": {
			description: "Creates a new role",
			funcTask:    t.createNewRoleTask,
		},
	}

	var names []string
	for k := range mapDetails {
		names = append(names, k)
	}
	slices.Sort(names)

	return names, mapDetails
}

// setUp initializes the application configuration, establishes a database connection,
// creates a new service store, and returns a new task instance with these dependencies.
// It returns an error if any step in the setup process fails.
func setUp() (*task, error) {
	config, err := app.LoadConfig()
	if err != nil {
		return nil, err
	}

	pool, err := db.Connect(config)
	if err != nil {
		return nil, err
	}

	store, err := service.New(config, pool)
	if err != nil {
		return nil, err
	}

	return &task{
		config:       config,
		serviceStore: store,
		reader:       os.Stdin,
	}, nil
}
