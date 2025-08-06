// Package db provides utilities for creating and managing a PostgreSQL database connection pool.
package db

import (
	"context"
	"fmt"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect establishes a connection to the database using the provided configuration,
// sets connection pool parameters, pings the database to verify connectivity,
// and returns a pgxpool.Pool instance or an error.
func Connect(config *app.Config) (*pgxpool.Pool, error) {
	ctx := context.Background()

	poolConfig, err := pgxpool.ParseConfig(config.DBURL)
	if err != nil {
		return nil, errs.WrapErrorWithMessage("failed to parse database url", err)
	}

	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.MaxConnLifetime = config.MaxConnLifetime

	timeout := fmt.Sprintf("%v", app.DefaultTimeout.Milliseconds())
	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = timeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, errs.WrapErrorWithMessage("unable to create database pool", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, errs.WrapErrorWithMessage("failed to ping database", err)
	}

	return pool, nil
}
