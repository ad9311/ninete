package db_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/db"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	config := app.FactoryConfig(t)
	pool := db.FactoryDBPool(t, config)
	defer pool.Close()

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_ping",
			func(t *testing.T) {
				ctx := context.Background()
				err := pool.Ping(ctx)
				require.Nil(t, err)
			},
		},
		{
			"should_apply_statement_timeout",
			func(t *testing.T) {
				runtimeParams := pool.Config().ConnConfig.RuntimeParams
				appTimout := strconv.FormatInt(app.DefaultTimeout.Milliseconds(), 10)
				require.Equal(t, appTimout, runtimeParams["statement_timeout"])
			},
		},
		{
			"should_apply_config",
			func(t *testing.T) {
				dbConfig := pool.Config()
				require.Equal(t, config.MaxConns, dbConfig.MaxConns)
				require.Equal(t, config.MinConns, dbConfig.MinConns)
				require.Equal(t, config.MaxConnIdleTime, dbConfig.MaxConnIdleTime)
				require.Equal(t, config.MaxConnLifetime, dbConfig.MaxConnLifetime)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
