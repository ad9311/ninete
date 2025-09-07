package task

import (
	"context"
)

// pingDatabaseTask pings the database and prints the result to the console.
func (t *task) pingDatabaseTask() error {
	ctx := context.Background()

	t.logger.Log("Pinging the database...")

	if err := t.serviceStore.PingDB(ctx); err != nil {
		return err
	}

	t.logger.Log("Database UP!")

	return nil
}
