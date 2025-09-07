package db_test

import (
	"context"
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
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
