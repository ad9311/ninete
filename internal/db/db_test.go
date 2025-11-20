package db_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	_, err := prog.Load()
	require.NoError(t, err)

	sqlDB, err := db.Open()
	require.NoError(t, err)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			"should_set_the_correct_max_connections",
			func(t *testing.T) {
				maxOpenConns := os.Getenv("MAX_OPEN_CONNS")
				if maxOpenConns == "" {
					maxOpenConns = strconv.Itoa(db.DefaultMaxOpenConns)
				}

				value := strconv.Itoa(sqlDB.Stats().MaxOpenConnections)
				require.Equal(t, maxOpenConns, value)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}

	t.Cleanup(func() {
		err := sqlDB.Close()
		require.NoError(t, err)
	})
}
