// Package main for running maintenance tasks
package main

import (
	"context"
	"os"

	"github.com/ad9311/ninete/internal/cmd"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/task"
)

func main() {
	tc, err := task.New()
	if err != nil {
		prog.QuickLogger().Error(err)
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
		{
			Name:        "create_expenses_from_recurrent",
			Description: "Create expenses from recurrent expenses",
			Run: func(_ []string) error {
				return tc.CreateExpensesFromRecurrent(context.Background())
			},
		},
	})
	if err != nil {
		tc.App.Logger.Error(err)
	}

	os.Exit(code)
}
