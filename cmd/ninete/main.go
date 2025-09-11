// Package main
package main

import (
	"os"

	"github.com/ad9311/ninete/internal/conf"
	"github.com/ad9311/ninete/internal/csl"
	"github.com/ad9311/ninete/internal/db"
)

func main() {
	l, err := csl.New(os.Stdout, os.Stderr)
	if err != nil {
		os.Exit(1)
	}

	env, err := conf.Load()
	if err != nil {
		l.Error("%v", err)
	}

	conn, err := db.Open(env)
	if err != nil {
		l.Error("%v", err)
	}
	defer conn.Close()

	l.Log("No errors...")
}
