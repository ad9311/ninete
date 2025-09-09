package db

import (
	"testing"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FactoryDBPool creates a database connection pool for use in tests and fails
// the test immediately if pool creation fails.
func FactoryDBPool(t *testing.T, config *app.Config) *pgxpool.Pool {
	t.Helper()

	pool, err := Connect(config)
	if err != nil {
		t.Fatalf("failed to create database pool: %v", err)
	}

	return pool
}
