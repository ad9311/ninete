// Package main
package main

import (
	"fmt"

	"github.com/ad9311/ninete/internal/conf"
	"github.com/ad9311/ninete/internal/db"
)

func main() {
	ac, _ := conf.Load()
	sqlDB, err := db.Open(ac.DBConf)
	if err != nil {
		fmt.Printf("%v", err)
	}

	fmt.Println(sqlDB.Ping())

	fmt.Println(sqlDB.Stats().MaxOpenConnections)

	fmt.Println(sqlDB.Close())
}
