#!/bin/bash
set -euo pipefail

SCRIPTS="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ENV_FILE="/etc/ninete/env"

APP_BIN_DIR="/srv/ninete/bin"
ARCHIVE_DIR="$APP_BIN_DIR/archive"

BUILT_BIN_APP_PATH="$APP_BIN_DIR/ninete"
MAIN_APP_PATH="/usr/local/bin/ninete"

SERVICE="ninete.service"

# Set between `systemctl stop` and `systemctl start` so the EXIT trap can bring
# the service back if anything in between fails under `set -e`.
SERVICE_STOPPED=0

# Must match deploy.sh: a rollback installs the whole set, not just the web
# binary, so cron and migrate.sh do not end up running newer code than the app.
ARCHIVED_BINARIES=(ninete migrate task)

TARGET=""
RESTORE_DB=0
SNAPSHOT=""
FORCE=0

usage() {
    cat <<USAGE
Usage: rollback.sh                             list archived versions
       rollback.sh <version>                   reinstall that version's binaries
       rollback.sh <version> --with-database   also restore the pre-deploy snapshot
                             [--snapshot PATH] restore a specific snapshot file
                             [--force]         install despite a schema mismatch

Rolling back the binaries alone is safe and loses nothing. --with-database
replaces the live database and loses every write made since the snapshot was
taken; it stops the service, which prompts for your sudo password.
USAGE
}

# Loads DATABASE_URL and friends the same way migrate.sh does. Only needed for
# the database paths — a binary-only rollback works without it.
load_env() {
    set -o allexport
    # shellcheck disable=SC1090,SC1091
    source "$ENV_FILE"
    set +o allexport
}

list_versions() {
    local installed dir version

    installed="$("$MAIN_APP_PATH" version 2>/dev/null || echo "unknown")"

    echo "==> Installed: $installed"
    echo "==> Archived versions:"

    if [ ! -d "$ARCHIVE_DIR" ] || [ -z "$(ls -A "$ARCHIVE_DIR" 2>/dev/null)" ]; then
        echo "    (none — only deploys made after archiving was added appear here)"
        return 0
    fi

    # Parsing `ls` is safe here and nowhere else: these names are `git describe`
    # output, which cannot contain a space or a newline. Newest first, so the
    # most likely rollback target is at the top.
    # shellcheck disable=SC2012,SC2045
    while IFS= read -r dir; do
        [ -n "$dir" ] || continue
        version="$(basename "$dir")"
        echo "    $version    schema $("$dir/migrate" schema-version 2>/dev/null || echo unknown)"
    done < <(ls -dt "$ARCHIVE_DIR"/*/ 2>/dev/null)
}

list_snapshots() {
    local dir="$1"

    # shellcheck disable=SC2012
    ls -t "$dir"/snapshot-*.db 2>/dev/null || true
}

# Refuses a rollback whose code predates the schema in the database. The archived
# migrate binary is asked what schema it carries, so the answer comes from the
# artifact being installed rather than from bookkeeping that could have drifted.
check_schema() {
    local dest="$1" archived live

    archived="$("$dest/migrate" schema-version 2>/dev/null || echo "")"
    live="$("$SCRIPTS/migrate.sh" db-version 2>/dev/null | grep -E "^[0-9]+$" | tail -n 1 || echo "")"

    if [ -z "$archived" ] || [ -z "$live" ]; then
        echo "    WARNING: could not compare schema versions (archived='$archived' live='$live')"
        return 0
    fi

    echo "    schema: target carries $archived, database is at $live"

    if [ "$live" -le "$archived" ]; then
        return 0
    fi

    # --with-database is the documented cure for this exact mismatch: the
    # snapshot it restores predates the migration that moved the database past
    # the target. Refusing here would make the advice impossible to follow.
    if [ "$RESTORE_DB" -eq 1 ]; then
        echo "    NOTE: the database has migrated past this version, but the snapshot"
        echo "          about to be restored predates that migration."

        return 0
    fi

    echo "    ERROR: the database has migrated past this version."
    echo "           Its code has never seen schema $live and may fail to read it."
    echo "           Re-run with --with-database to restore the pre-deploy snapshot,"
    echo "           or with --force if you are certain the old code tolerates it."

    [ "$FORCE" -eq 1 ] || exit 1

    echo "    (--force given, continuing)"
}

# Armed for the whole run. Between `systemctl stop` and `systemctl start` any
# failure aborts under `set -e`, and without this the app is left down with no
# hint as to why. Silent unless the service was actually stopped.
#
# shellcheck disable=SC2329 # invoked by the EXIT trap in main(), not by name
restart_on_failure() {
    [ "$SERVICE_STOPPED" -eq 1 ] || return 0

    echo "ERROR: the rollback failed while $SERVICE was stopped; starting it again."
    echo "       The database may be mid-restore — check it before trusting it."
    sudo systemctl start "$SERVICE" || true
}

confirm() {
    local prompt="$1" reply

    if [ ! -t 0 ]; then
        echo "ERROR: rollback.sh needs a terminal to confirm. Run it interactively."
        exit 1
    fi

    read -r -p "$prompt [y/N] " reply
    case "$reply" in
        [yY] | [yY][eE][sS]) return 0 ;;
        *)
            echo "==> Rollback aborted."
            exit 1
            ;;
    esac
}

# Copies the archived binaries into the staging directory and installs the web
# binary from there. The four privileged commands are byte-identical to the ones
# in deploy.sh because the host's sudoers policy matches them literally.
install_binaries() {
    local dest="$1" binary

    echo "==> Installing $TARGET binaries..."
    for binary in "${ARCHIVED_BINARIES[@]}"; do
        cp "$dest/$binary" "$APP_BIN_DIR/$binary"
    done

    sudo cp "$BUILT_BIN_APP_PATH" "$MAIN_APP_PATH"
    sudo chown root:root "$MAIN_APP_PATH"
    sudo chmod 755 "$MAIN_APP_PATH"
}

# Replaces the live database with a snapshot. Three things make this safe, and
# all three are load-bearing:
#   - the service is stopped, not restarted: overwriting the file under a running
#     process corrupts its view of the database;
#   - a fresh snapshot is taken first, so the restore itself is reversible;
#   - the -wal and -shm sidecars are removed, because a stale WAL replaying onto
#     a restored database is exactly the corruption this is meant to avoid.
restore_database() {
    local snapshot="$1" magic staged

    if [ ! -f "$snapshot" ]; then
        echo "ERROR: snapshot '$snapshot' does not exist."
        exit 1
    fi

    magic="$(head -c 15 "$snapshot" 2>/dev/null || true)"
    if [ "$magic" != "SQLite format 3" ]; then
        echo "ERROR: '$snapshot' is not a SQLite database."
        exit 1
    fi

    # Staged outside the snapshot directory before anything else runs: the
    # safety snapshot below prunes to the retention limit, and would otherwise
    # be able to delete the very file being restored (reachable with
    # --snapshot pointing at one of the older kept snapshots).
    staged="$DATABASE_URL.restore"
    cp "$snapshot" "$staged"

    echo "==> Stopping $SERVICE (this asks for your sudo password)..."
    sudo systemctl stop "$SERVICE"
    SERVICE_STOPPED=1

    echo "==> Snapshotting the current database before replacing it..."
    "$SCRIPTS/migrate.sh" snapshot

    echo "==> Restoring $snapshot over $DATABASE_URL..."
    rm -f "$DATABASE_URL-wal" "$DATABASE_URL-shm"
    # cp, not mv: the destination inode keeps the ownership and mode the
    # service unit expects.
    cp "$staged" "$DATABASE_URL"
    rm -f "$staged"
}

parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --with-database) RESTORE_DB=1 ;;
            --snapshot)
                shift
                SNAPSHOT="${1:-}"
                if [ -z "$SNAPSHOT" ]; then
                    echo "ERROR: --snapshot needs a path."
                    exit 1
                fi
                RESTORE_DB=1
                ;;
            --force) FORCE=1 ;;
            -h | --help)
                usage
                exit 0
                ;;
            -*)
                echo "ERROR: unknown option '$1'."
                usage
                exit 1
                ;;
            *)
                if [ -n "$TARGET" ]; then
                    echo "ERROR: only one version can be given."
                    exit 1
                fi
                TARGET="$1"
                ;;
        esac
        shift
    done
}

main() {
    local dest snapshot

    if [ "$EUID" -eq 0 ]; then
        echo "ERROR: Do not run this script as root."
        exit 1
    fi

    trap restart_on_failure EXIT

    parse_args "$@"

    if [ -z "$TARGET" ]; then
        list_versions
        echo
        echo "Pass a version to roll back to it. rollback.sh --help for options."
        exit 0
    fi

    dest="$ARCHIVE_DIR/$TARGET"
    if [ ! -d "$dest" ]; then
        echo "ERROR: no archived build for '$TARGET'."
        echo
        list_versions
        exit 1
    fi

    echo "==> Rolling back to $TARGET"
    echo "    installed: $("$MAIN_APP_PATH" version 2>/dev/null || echo unknown)"

    check_schema "$dest"

    if [ "$RESTORE_DB" -eq 1 ]; then
        load_env

        snapshot="$SNAPSHOT"
        if [ -z "$snapshot" ]; then
            # Must match snapshotDir() in internal/db/snapshot.go, which honours
            # SNAPSHOT_DIR before falling back to a directory beside the database.
            snapshot="$(list_snapshots "${SNAPSHOT_DIR:-$(dirname "$DATABASE_URL")/snapshots}" | head -n 1)"
        fi

        if [ -z "$snapshot" ]; then
            echo "ERROR: no snapshot found to restore."
            exit 1
        fi

        echo
        echo "    WARNING: restoring $(basename "$snapshot") DISCARDS every change"
        echo "             written to the database since that snapshot was taken."
        confirm "    Restore the database and roll back to $TARGET?"

        restore_database "$snapshot"
        install_binaries "$dest"

        echo "==> Starting $SERVICE..."
        sudo systemctl start "$SERVICE"
        SERVICE_STOPPED=0
    else
        confirm "    Reinstall $TARGET binaries and restart?"

        install_binaries "$dest"

        echo "==> Restarting $SERVICE..."
        sudo systemctl restart "$SERVICE"
    fi

    echo "==> Rolled back to $("$MAIN_APP_PATH" version 2>/dev/null || echo unknown)"
}

# `exit` shares this line deliberately. Bash parses one command at a time and
# seeks back to just past it before executing, so returning from a `main` that
# rewrote this file would resume at a stale offset in the new bytes. An `exit`
# on its own line lives in the old bytes and is never read.
main "$@"; exit
