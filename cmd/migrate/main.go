// Package main for migrations
package main

import (
	"fmt"

	"github.com/ad9311/ninete/internal/db"
)

func main() {
	if err := db.RunMigrationsUp(); err != nil {
		fmt.Println(err)
	}
}
