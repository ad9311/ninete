package task

import (
	"context"
	"fmt"
)

// pingDatabaseTask pings the database and prints the result to the console.
func (t *task) pingDatabaseTask() error {
	ctx := context.Background()

	fmt.Println("\nPinging the database...")

	if err := t.serviceStore.PingDB(ctx); err != nil {
		return err
	}

	fmt.Println("Database UP!")

	return nil
}
