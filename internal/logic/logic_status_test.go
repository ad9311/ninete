package logic_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestAppStatus(t *testing.T) {
	f := testhelper.NewFactory(t)

	cases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "should_return_app_status",
			testFunc: func(t *testing.T) {
				stats, err := f.Store.AppStatus()
				require.NoError(t, err)

				require.Equal(t, "TODO", stats.Version)
				require.Equal(t, os.Getenv("ENV"), stats.ENV)

				require.GreaterOrEqual(t, stats.DBConnStats.MaxOpenConnections, 0)
				require.GreaterOrEqual(t, stats.DBConnStats.IdleConnections, 0)
				require.GreaterOrEqual(t, stats.DBConnStats.InUseConnections, 0)

				require.NotEmpty(t, stats.MemStats.Allocated)
				require.True(t, strings.HasSuffix(stats.MemStats.Allocated, "MiB"))
				require.NotEmpty(t, stats.MemStats.System)
				require.True(t, strings.HasSuffix(stats.MemStats.System, "MiB"))
			},
		},
		{
			name: "should_fail_when_database_ping_fails",
			testFunc: func(t *testing.T) {
				f.CloseDB(t)

				_, err := f.Store.AppStatus()
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to ping database")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.testFunc)
	}
}
