package db

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/pressly/goose/v3"
)

const (
	// snapshotDirName is created beside the database file unless SNAPSHOT_DIR
	// overrides it. Beside the database is the one place the deploy account is
	// certain to be able to write.
	snapshotDirName = "snapshots"

	snapshotPrefix    = "snapshot-"
	snapshotExtension = ".db"

	// snapshotRetention is how many snapshots survive a prune. A rollback that
	// reaches further back than the last few deploys wants the off-site backup,
	// not this.
	snapshotRetention = 5

	// maxSnapshotNameAttempts caps the search for a free filename within one
	// second. Reaching it means something is looping, not that a name is taken.
	maxSnapshotNameAttempts = 50

	snapshotDirPerm = 0o750
)

// SnapshotDatabase writes a consistent copy of the database beside it and prunes
// older copies. It is the pre-deploy safety net: taken before migrations run, it
// is what a rollback restores when a migration turns out to be wrong.
//
// VACUUM INTO, not a file copy: copying a database with a live WAL can capture a
// torn state. The snapshot is also compacted, so it is normally smaller than the
// source.
func SnapshotDatabase() error {
	// setUpMigrator runs first because it calls prog.Load, which is what puts
	// .env into the environment outside production. Reading DATABASE_URL before
	// it would see nothing in development.
	app, sqlDB, err := setUpMigrator()
	if err != nil {
		return err
	}
	defer closeSQLDB(app, sqlDB)

	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		return fmt.Errorf("'DATABASE_URL' %w", prog.ErrEnvNoTSet)
	}

	dir, err := snapshotDir(dbPath)
	if err != nil {
		return err
	}

	path, err := uniqueSnapshotPath(dir, prog.Version)
	if err != nil {
		return err
	}

	// VACUUM does not accept bind parameters, so the path is inlined. Doubling
	// any single quote keeps a path containing one from ending the literal. The
	// path is built here from SNAPSHOT_DIR and the build stamp, never from a
	// request.
	quoted := strings.ReplaceAll(path, "'", "''")
	//nolint:gosec // G202: VACUUM INTO takes no bind parameters; the path is internal and quote-escaped
	if _, err := sqlDB.Exec("VACUUM INTO '" + quoted + "'"); err != nil {
		return fmt.Errorf("%w: %w", ErrSnapshotFailed, err)
	}

	// Printed rather than logged: deploy.sh reports the path it just created.
	fmt.Println(path)

	// Pruning is housekeeping and must not fail the command: the snapshot is
	// already on disk, and rollback.sh takes one with the service stopped, where
	// a non-zero exit aborts the restore over a file it could not delete.
	pruned, err := pruneSnapshots(dir)
	if err != nil {
		app.Logger.Errorf("%v", err)
	} else if pruned > 0 {
		app.Logger.Logf("Pruned %d old snapshot(s), keeping %d", pruned, snapshotRetention)
	}

	return nil
}

// PrintDBVersion prints the migration version the database is currently at, as a
// bare number so a shell script can capture it.
func PrintDBVersion() error {
	app, sqlDB, err := setUpMigrator()
	if err != nil {
		return err
	}
	defer closeSQLDB(app, sqlDB)

	version, err := goose.GetDBVersion(sqlDB)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDBVersionRead, err)
	}

	fmt.Println(version)

	return nil
}

// PrintSchemaVersion prints the newest migration this binary carries, as a bare
// number. It reads only the embedded migrations, so it needs neither ENV nor a
// database — which is the point: rollback.sh asks an *archived* binary what
// schema it expects before installing it, on a host whose database may be ahead.
func PrintSchemaVersion() error {
	version, err := schemaVersion()
	if err != nil {
		return err
	}

	fmt.Println(version)

	return nil
}

func schemaVersion() (int64, error) {
	entries, err := fs.ReadDir(embededMigrations, migrationsPath)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrSchemaVersionRead, err)
	}

	var newest int64
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		prefix, _, found := strings.Cut(entry.Name(), "_")
		if !found {
			continue
		}

		version, err := strconv.ParseInt(prefix, 10, 64)
		if err != nil {
			continue
		}

		if version > newest {
			newest = version
		}
	}

	if newest == 0 {
		return 0, ErrSchemaVersionRead
	}

	return newest, nil
}

func snapshotDir(dbPath string) (string, error) {
	dir := os.Getenv("SNAPSHOT_DIR")
	if dir == "" {
		dir = filepath.Join(filepath.Dir(dbPath), snapshotDirName)
	}

	// The path is tainted by SNAPSHOT_DIR, which only the operator sets: production
	// config arrives through the systemd EnvironmentFile, and anyone who can set it
	// already runs this process.
	// #nosec G703 -- the only untrusted input here would be a hostile process environment.
	if err := os.MkdirAll(dir, snapshotDirPerm); err != nil {
		return "", fmt.Errorf("%w: %w", ErrSnapshotDir, err)
	}

	return dir, nil
}

// snapshotName stamps the build being deployed into the filename. At pre-deploy
// time the binary taking the snapshot is the new one, so the name reads as "the
// database as it stood before <version> was deployed".
func snapshotName(version string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '.', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, version)

	return snapshotPrefix + safe + "-" + time.Now().UTC().Format("20060102T150405Z")
}

// uniqueSnapshotPath resolves the collision between two snapshots taken in the
// same second — VACUUM INTO refuses to overwrite, and the name is only
// second-resolution. A deploy never hits this; a script run twice by hand does.
func uniqueSnapshotPath(dir, version string) (string, error) {
	base := snapshotName(version)

	path := filepath.Join(dir, base+snapshotExtension)
	for i := 2; ; i++ {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return path, nil
		}

		// Anything other than "not there" is a real failure — an unreadable
		// directory, say. Trying 50 more names would only bury it.
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrSnapshotFailed, err)
		}

		if i > maxSnapshotNameAttempts {
			return "", ErrSnapshotNameTaken
		}

		path = filepath.Join(dir, base+"-"+strconv.Itoa(i)+snapshotExtension)
	}
}

// pruneSnapshots keeps the newest snapshotRetention files by modification time.
// Names carry a version as well as a timestamp, so they do not sort
// chronologically and mtime is the only ordering that means anything.
func pruneSnapshots(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrSnapshotDir, err)
	}

	type snapshot struct {
		path    string
		modTime time.Time
	}

	var snapshots []snapshot
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, snapshotPrefix) || !strings.HasSuffix(name, snapshotExtension) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		snapshots = append(snapshots, snapshot{path: filepath.Join(dir, name), modTime: info.ModTime()})
	}

	if len(snapshots) <= snapshotRetention {
		return 0, nil
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].modTime.After(snapshots[j].modTime)
	})

	var pruned int
	for _, s := range snapshots[snapshotRetention:] {
		if err := os.Remove(s.path); err != nil {
			return pruned, fmt.Errorf("%w: %w", ErrSnapshotPrune, err)
		}
		pruned++
	}

	return pruned, nil
}
