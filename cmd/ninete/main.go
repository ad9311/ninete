// Package main
package main

import (
	"log"

	"github.com/ad9311/ninete/internal/conf"
	"github.com/ad9311/ninete/internal/db"
)

func main() {
	ac, err := conf.Load()
	if err != nil {
		log.Println(err)
	}

	conn, err := db.Open(ac.DBConf)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	if err := conn.DB.Ping(); err != nil {
		log.Println(err)
	}
}
