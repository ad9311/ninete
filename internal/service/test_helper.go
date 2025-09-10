package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/console"
	"github.com/ad9311/go-api-base/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RunWithIsolatedSchema sets up a unique database schema for the test run, runs migrations into it,
// sets the DATABASE_URL environment variable to use the schema, runs the tests, and then cleans up
// by dropping the schema. This ensures tests run in isolation without affecting other data.
func RunWithIsolatedSchema(m *testing.M, packageName string) int {
	config, err := app.LoadConfig()
	if err != nil {
		console.NewError("failed to load config: %v", err)

		return 2
	}

	schema := "test_" + packageName + "_" + randHex(6) // e.g. test_service_a1b2c3
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Connect(config)
	if err != nil {
		console.NewError("failed to open database pool: %v", err)

		return 2
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		console.NewError("failed to set schema: %v", err)

		return 2
	}

	testDSN := appendSearchPath(config.DBConfig.URL, schema)

	if err := os.Setenv("DATABASE_URL", testDSN); err != nil {
		console.NewError("failed to set DATABASE_URL: %v", err)

		return 1
	}
	defer resetDBURLEnv()

	if err := runMigrationsIntoSchema(ctx, pool, schema); err != nil {
		console.NewError("failed to run migrations: %v", err)
		err = dropSchema(ctx, pool, schema)
		if err != nil {
			console.NewError("failed to drop schema after migrations failed: %v", err)
		}

		return 1
	}

	code := m.Run()

	if err := dropSchema(ctx, pool, schema); err != nil {
		console.NewError("failed to drop schema: %v", err)
	}

	return code
}

// runMigrationsIntoSchema sets the PostgreSQL search_path to the specified schema and runs database migrations.
// It executes a SQL command to change the search_path, then calls db.RunMigrationsUp to apply migrations.
// Returns an error if setting the search_path or running migrations fails.
func runMigrationsIntoSchema(ctx context.Context, pool *pgxpool.Pool, schema string) error {
	_, err := pool.Exec(ctx, "SET search_path = "+schema+",public")
	if err != nil {
		return err
	}

	return db.RunMigrationsUp([]string{})
}

// dropSchema drops the specified PostgreSQL schema and all its dependent objects from the database.
// It executes a "DROP SCHEMA IF EXISTS <schema> CASCADE" statement using the provided pgxpool.Pool.
// Returns an error if the operation fails.
func dropSchema(ctx context.Context, pool *pgxpool.Pool, schema string) error {
	_, err := pool.Exec(ctx, "DROP SCHEMA IF EXISTS "+schema+" CASCADE")

	return err
}

// randHex generates a random hexadecimal string of length 2*n,
// where n is the number of random bytes generated.
// It uses crypto/rand for secure random byte generation.
func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}

// appendSearchPath appends a "search_path" query parameter with the given schema
// to the provided URL. If the URL already contains query parameters, the new
// parameter is appended using '&'; otherwise, it is added using '?'.
// Returns the modified URL as a string.
func appendSearchPath(url, schema string) string {
	searchPath := "search_path=" + schema
	if strings.Contains(url, "?") {
		return url + "&" + searchPath
	}

	return url + "?" + searchPath
}

// resetDBURLEnv resets the "DATABASE_URL" environment variable to an empty string.
// If setting the environment variable fails, it logs an error message using console.NewError.
func resetDBURLEnv() {
	err := os.Setenv("DATABASE_URL", "")
	if err != nil {
		console.NewError("failed to reset DATABASE_URL: %v", err)
	}
}
