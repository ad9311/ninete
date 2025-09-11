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
		l.Error("%v", err)
	}

	ac, err := conf.Load()
	if err != nil {
		l.Error("%v", err)
	}

	conn, err := db.Open(ac.DBConf)
	if err != nil {
		l.Error("%v", err)
	}
	defer conn.Close()

	l.Log("No errors...")
}
