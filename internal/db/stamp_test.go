package db_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/stretchr/testify/require"
)

// openStamped opens dbPath under the given ENV. The caller closes the pool
// before opening the same file again, so each open sees the stamp the previous
// one left behind.
func openStamped(t *testing.T, env, dbPath string) (*sql.DB, error) {
	t.Helper()

	t.Setenv("ENV", env)
	t.Setenv("DATABASE_URL", dbPath)

	sqlDB, err := db.Open()
	if sqlDB != nil {
		t.Cleanup(func() {
			_ = sqlDB.Close()
		})
	}

	return sqlDB, err
}

func readAppID(t *testing.T, sqlDB *sql.DB) int64 {
	t.Helper()

	var appID int64
	require.NoError(t, sqlDB.QueryRow(`PRAGMA application_id`).Scan(&appID))

	return appID
}

func TestEnvStamp(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "claims an unstamped database for the current environment",
			fn: func(t *testing.T) {
				dbPath := filepath.Join(t.TempDir(), "stamp_claim.db")

				sqlDB, err := openStamped(t, prog.ENVTest, dbPath)
				require.NoError(t, err)
				require.NotZero(t, readAppID(t, sqlDB), "database left unstamped")
			},
		},
		{
			name: "reopening under the same environment succeeds",
			fn: func(t *testing.T) {
				dbPath := filepath.Join(t.TempDir(), "stamp_reopen.db")

				first, err := openStamped(t, prog.ENVTest, dbPath)
				require.NoError(t, err)
				stamp := readAppID(t, first)
				require.NoError(t, first.Close())

				second, err := openStamped(t, prog.ENVTest, dbPath)
				require.NoError(t, err)
				require.Equal(t, stamp, readAppID(t, second), "stamp changed on reopen")
			},
		},
		{
			// Reproduction: before the stamp, a development command pointed at
			// the production database opened it happily and seeded dev users
			// into it.
			name: "refuses a database stamped for another environment",
			fn: func(t *testing.T) {
				dbPath := filepath.Join(t.TempDir(), "stamp_mismatch.db")

				owner, err := openStamped(t, prog.ENVProduction, dbPath)
				require.NoError(t, err)
				require.NoError(t, owner.Close())

				_, err = openStamped(t, prog.ENVDevelopment, dbPath)
				require.ErrorIs(t, err, db.ErrEnvStampMismatch)
			},
		},
		{
			name: "refuses an unknown environment",
			fn: func(t *testing.T) {
				dbPath := filepath.Join(t.TempDir(), "stamp_unknown.db")

				_, err := openStamped(t, "staging", dbPath)
				require.True(
					t,
					errors.Is(err, db.ErrUnknownEnvStamp),
					"expected an unknown environment error, got %v", err,
				)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}
