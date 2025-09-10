// Package main
package main

import (
	"fmt"

	"github.com/ad9311/ninete/internal/conf"
)

func main() {
	ac, _ := conf.Load()
	fmt.Printf("%+v\n", ac)
}
