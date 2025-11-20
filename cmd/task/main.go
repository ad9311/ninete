// Package main for running maintenance tasks
package main

import (
	"os"

	"github.com/ad9311/ninete/internal/cmd"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/task"
)

func main() {
	code, err := cmd.Run(os.Args[0], []*cmd.Command{
		{
			Name:        "dev",
			Description: "Run test code",
			Run: func(_ []string) error {
				return task.RunDev()
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
