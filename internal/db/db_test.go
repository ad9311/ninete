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
	require.Nil(t, err)

	sqlDB, err := db.Open()
	require.Nil(t, err)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			"should_set_the_correct_max_connections",
			func(t *testing.T) {
				maxOpenConns := os.Getenv("MAX_OPEN_CONNS")
				value := strconv.Itoa(sqlDB.Stats().MaxOpenConnections)
				require.Equal(t, maxOpenConns, value)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
