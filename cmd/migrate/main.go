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
			Run:         db.RunMigrationsUp,
		},
		{
			Name:        "down",
			Description: "Run one migration down",
			Run:         db.RunMigrationsDown,
		},
		{
			Name:        "create",
			Description: "Create a new migration file",
			Run:         db.CreateMigration,
		},
		{
			Name:        "status",
			Description: "Print migrations status",
			Run:         db.PrintStatus,
		},
		{
			Name:        "seed",
			Description: "Run database seeds",
			Run:         db.RunSeeds,
		},
	})
	if err != nil {
		prog.QuickLogger().Errorf("%v", err)
	}

	os.Exit(code)
}
