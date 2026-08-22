package db_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/stretchr/testify/require"
)

// snapshotEnv points the snapshot machinery at a throwaway database and returns
// the directory snapshots land in.
func snapshotEnv(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "snapshot_test.db")

	t.Setenv("ENV", "test")
	t.Setenv("DATABASE_URL", dbPath)

	sqlDB, err := db.Open()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return filepath.Join(dir, "snapshots")
}

func snapshotFiles(t *testing.T, dir string) []os.DirEntry {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	return entries
}

func TestSnapshotDatabase(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "writes a snapshot beside the database",
			fn: func(t *testing.T) {
				dir := snapshotEnv(t)

				require.NoError(t, db.SnapshotDatabase())

				entries := snapshotFiles(t, dir)
				require.Len(t, entries, 1)
				require.True(t, strings.HasPrefix(entries[0].Name(), "snapshot-"))
				require.True(t, strings.HasSuffix(entries[0].Name(), ".db"))
			},
		},
		{
			// VACUUM INTO refuses to overwrite, and the name is only
			// second-resolution, so two snapshots in the same second must not
			// collide.
			name: "does not collide within the same second",
			fn: func(t *testing.T) {
				dir := snapshotEnv(t)

				require.NoError(t, db.SnapshotDatabase())
				require.NoError(t, db.SnapshotDatabase())

				require.Len(t, snapshotFiles(t, dir), 2)
			},
		},
		{
			name: "honours SNAPSHOT_DIR",
			fn: func(t *testing.T) {
				snapshotEnv(t)

				custom := filepath.Join(t.TempDir(), "elsewhere")
				t.Setenv("SNAPSHOT_DIR", custom)

				require.NoError(t, db.SnapshotDatabase())

				require.Len(t, snapshotFiles(t, custom), 1)
			},
		},
		{
			// The snapshot is what a rollback restores, so it has to be a
			// readable database, not merely a file that was created.
			name: "writes a readable database",
			fn: func(t *testing.T) {
				dir := snapshotEnv(t)

				require.NoError(t, db.SnapshotDatabase())

				entries := snapshotFiles(t, dir)
				require.Len(t, entries, 1)

				t.Setenv("DATABASE_URL", filepath.Join(dir, entries[0].Name()))

				restored, err := db.Open()
				require.NoError(t, err)
				require.NoError(t, restored.Ping())
				require.NoError(t, restored.Close())
			},
		},
		{
			name: "prunes down to the retention limit",
			fn: func(t *testing.T) {
				dir := snapshotEnv(t)

				for range 8 {
					require.NoError(t, db.SnapshotDatabase())
				}

				require.Len(t, snapshotFiles(t, dir), 5)
			},
		},
		{
			name: "leaves unrelated files in the directory alone",
			fn: func(t *testing.T) {
				dir := snapshotEnv(t)

				require.NoError(t, db.SnapshotDatabase())

				keep := filepath.Join(dir, "notes.txt")
				require.NoError(t, os.WriteFile(keep, []byte("keep me"), 0o600))

				for range 8 {
					require.NoError(t, db.SnapshotDatabase())
				}

				_, err := os.Stat(keep)
				require.NoError(t, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

func TestPrintSchemaVersion(t *testing.T) {
	// Reads only the embedded migrations, so it must work with no ENV and no
	// database — rollback.sh asks an archived binary this question on a host
	// whose database may be ahead of it.
	t.Setenv("ENV", "")
	t.Setenv("DATABASE_URL", "")

	require.NoError(t, db.PrintSchemaVersion())
}
