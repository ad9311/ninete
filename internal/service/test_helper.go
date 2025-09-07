package service

import (
	"context"
	"testing"
	"time"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/db"
)

const truncateQuery = `
  DELETE FROM user_roles;
	DELETE FROM roles;
	DELETE FROM refresh_tokens;
	DELETE FROM users
`

// RunTestsWithCleanUp runs all package tests and cleans the database
func RunTestsWithCleanUp(m *testing.M) int {
	config, err := app.LoadConfig()
	if err != nil {
		return 1
	}
	pool, err := db.Connect(config)
	if err != nil {
		return 1
	}
	defer pool.Close()

	code := m.Run()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second) // TODO
	defer cancel()

	_, err = pool.Exec(ctx, truncateQuery)
	if err != nil {
		return 1
	}

	return code
}
