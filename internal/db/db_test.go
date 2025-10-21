package db_test

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/stretchr/testify/require"
)

func OpenTestDB(url string) (*sql.DB, error) {
	var sqlDB *sql.DB

	if url == "" {
		return nil, fmt.Errorf("'DATABASE_URL' %w", prog.ErrEnvNoTSet)
	}

	if err := os.Setenv("DATABASE_URL", url); err != nil {
		return nil, fmt.Errorf("failed set DATABASE_URL env: %w", err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return sqlDB, nil
}

func TestOpen(t *testing.T) {
	_, err := prog.Load()
	require.NoError(t, err)

	sqlDB, err := db.Open()
	require.NoError(t, err)

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
