// Package main for migrations
package main

import (
	"os"

	"github.com/ad9311/ninete/internal/cmd"
	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/prog"
)

func main() {
	code, err := cmd.Run(os.Args[0], []*cmd.Command{
		{
			Name:        "up",
			Description: "Run all migrations",
			Run: func(_ []string) error {
				return db.RunMigrationsUp()
			},
		},
		{
			Name:        "down",
			Description: "Run one migration down",
			Run: func(_ []string) error {
				return db.RunMigrationsDown()
			},
		},
		{
			Name:        "create",
			Description: "Create a new migration file",
			Run:         db.CreateMigration,
		},
		{
			Name:        "status",
			Description: "Print migrations status",
			Run: func(_ []string) error {
				return db.PrintStatus()
			},
		},
	})
	if err != nil {
		prog.NewLogger(prog.LogOptions{
			EnableColor: true,
		}).Errorf("%v", err)
	}

	os.Exit(code)
}
