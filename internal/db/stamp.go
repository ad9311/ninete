package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/ad9311/ninete/internal/prog"
)

// Environment stamps written into the SQLite header's "application_id" field.
// The high three bytes spell "NIN" so a database belonging to another program
// is never mistaken for an unstamped one; the low byte identifies the
// environment that owns the file.
//
// The field is 32 bits wide. Values that do not fit are silently stored as
// zero, which is what the old "PRAGMA application_id = 0x6E696E657465" line in
// "init/database.sql" did, so keep these inside int32.
const (
	appIDProduction  = 0x4E494E01
	appIDDevelopment = 0x4E494E02
	appIDTest        = 0x4E494E03

	appIDUnstamped = 0
)

func envAppIDs() map[string]int64 {
	return map[string]int64{
		prog.ENVProduction:  appIDProduction,
		prog.ENVDevelopment: appIDDevelopment,
		prog.ENVTest:        appIDTest,
	}
}

func envForAppID(appID int64) string {
	for env, id := range envAppIDs() {
		if id == appID {
			return env
		}
	}

	return "unknown"
}

// verifyEnvStamp ties a database file to the environment that owns it. ENV and
// DATABASE_URL are independent variables, so nothing else stops a development
// command from opening the production database and seeding it with test users.
// An unstamped file is claimed for the current environment; a file stamped for
// another one fails the open before any statement runs.
func verifyEnvStamp(sqlDB *sql.DB) error {
	env := os.Getenv("ENV")

	want, ok := envAppIDs()[env]
	if !ok {
		return fmt.Errorf("%w: '%s'", ErrUnknownEnvStamp, env)
	}

	got, err := readAppID(sqlDB)
	if err != nil {
		return err
	}

	if got == appIDUnstamped {
		return writeAppID(sqlDB, want)
	}

	if got != want {
		return fmt.Errorf(
			"%w: database belongs to '%s', ENV is '%s'",
			ErrEnvStampMismatch, envForAppID(got), env,
		)
	}

	return nil
}

// StampDatabase claims an existing database for the current environment. It
// exists for databases created before stamping, which read back as unstamped:
// whichever command opens one first would otherwise claim it, so the
// production database must be stamped deliberately, once, with ENV=production.
func StampDatabase() error {
	app, err := prog.Load()
	if err != nil {
		return err
	}

	sqlDB, err := Open()
	if err != nil {
		return err
	}
	defer closeSQLDB(app, sqlDB)

	// Open has already claimed the file when it was unstamped, and refused to
	// return at all when it was stamped for another environment, so reaching
	// here means the stamp matches.
	appID, err := readAppID(sqlDB)
	if err != nil {
		return err
	}

	app.Logger.Logf("Database stamped for '%s' [application_id=%#x]", envForAppID(appID), appID)

	return nil
}

func readAppID(sqlDB *sql.DB) (int64, error) {
	var appID int64
	if err := sqlDB.QueryRow(`PRAGMA application_id`).Scan(&appID); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrEnvStampRead, err)
	}

	return appID, nil
}

func writeAppID(sqlDB *sql.DB, appID int64) error {
	// PRAGMA statements take no bind parameters. appID always comes from the
	// constants above, never from input.
	stmt := fmt.Sprintf(`PRAGMA application_id = %d`, appID)

	if _, err := sqlDB.Exec(stmt); err != nil {
		return fmt.Errorf("%w: %w", ErrEnvStampWrite, err)
	}

	return nil
}
