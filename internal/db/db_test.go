package db_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/stretchr/testify/require"
)

// openPool opens a pool allowed to hold several connections, so the tests can
// inspect connections other than the first one the driver hands out.
func openPool(t *testing.T) *sql.DB {
	t.Helper()

	t.Setenv("DATABASE_URL", filepath.Join(t.TempDir(), "pragma_test.db"))
	t.Setenv("MAX_OPEN_CONNS", "4")
	t.Setenv("MAX_IDLE_CONNS", "4")

	sqlDB, err := db.Open()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	return sqlDB
}

// grabConns holds n connections open at once, forcing the pool to create new
// ones instead of handing back the connection the startup PRAGMAs ran on.
func grabConns(ctx context.Context, t *testing.T, sqlDB *sql.DB, n int) []*sql.Conn {
	t.Helper()

	conns := make([]*sql.Conn, 0, n)

	for range n {
		conn, err := sqlDB.Conn(ctx)
		require.NoError(t, err)

		conns = append(conns, conn)
	}

	t.Cleanup(func() {
		for _, conn := range conns {
			require.NoError(t, conn.Close())
		}
	})

	return conns
}

func TestConnectionPragmas(t *testing.T) {
	const connCount = 4

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "applies connection pragmas to every connection in the pool",
			fn: func(t *testing.T) {
				ctx := t.Context()
				sqlDB := openPool(t)

				for i, conn := range grabConns(ctx, t, sqlDB, connCount) {
					var foreignKeys int
					require.NoError(
						t,
						conn.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&foreignKeys),
						"connection %d", i,
					)
					require.Equal(t, 1, foreignKeys, "foreign_keys off on connection %d", i)

					var busyTimeout int
					require.NoError(
						t,
						conn.QueryRowContext(ctx, `PRAGMA busy_timeout`).Scan(&busyTimeout),
						"connection %d", i,
					)
					require.Equal(t, 5000, busyTimeout, "busy_timeout unset on connection %d", i)
				}
			},
		},
		{
			name: "cascades deletes on a connection other than the first",
			fn: func(t *testing.T) {
				ctx := t.Context()
				sqlDB := openPool(t)

				_, err := sqlDB.ExecContext(ctx, `
CREATE TABLE "parents" ("id" INTEGER PRIMARY KEY NOT NULL);
CREATE TABLE "children" (
  "id"        INTEGER PRIMARY KEY NOT NULL,
  "parent_id" INTEGER NOT NULL REFERENCES "parents"("id") ON DELETE CASCADE
);
INSERT INTO "parents" ("id") VALUES (1);
INSERT INTO "children" ("id", "parent_id") VALUES (1, 1);`)
				require.NoError(t, err)

				conns := grabConns(ctx, t, sqlDB, connCount)
				last := conns[len(conns)-1]

				_, err = last.ExecContext(ctx, `DELETE FROM "parents" WHERE "id" = 1`)
				require.NoError(t, err)

				var orphans int
				require.NoError(t, last.QueryRowContext(ctx, `SELECT COUNT(*) FROM "children"`).Scan(&orphans))
				require.Zero(t, orphans, "child row survived the cascade")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}
