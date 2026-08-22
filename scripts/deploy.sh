#!/bin/bash
set -euo pipefail

SCRIPTS="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd -P "$SCRIPTS/.." && pwd)"

APP_BIN_DIR="/srv/ninete/bin"
ARCHIVE_DIR="$APP_BIN_DIR/archive"

BUILT_BIN_APP_PATH="$APP_BIN_DIR/ninete"
MAIN_APP_PATH="/usr/local/bin/ninete"

# Binaries kept per archived version. All three, not just the web binary: cron
# runs task and migrate.sh runs migrate, so rolling the code back without them
# leaves a mismatched set.
ARCHIVED_BINARIES=(ninete migrate task)

# How many past versions rollback.sh can reach. Further back than this is a job
# for the off-site backup and a rebuild.
ARCHIVE_RETENTION=3

# Set by --yes/-y to skip the confirmation prompt.
ASSUME_YES=0

# Resolved after pull.sh, since the version describes the code it just fetched.
VERSION=""

# Reports what this deploy would stamp into the binaries, and stops for
# confirmation when that stamp says something is off: an untagged HEAD, or a
# dirty checkout. Neither is an error — untagged deploys are the normal case and
# `git describe` still yields an exact identity — so this only asks, it never
# refuses on its own.
#
# It runs after pull.sh, because the version is a property of the code that was
# just fetched, and before build.sh and migrate.sh, so declining costs nothing
# beyond an advanced checkout.
confirm_version() {
    local warnings=() reply

    echo "==> Deploying version $VERSION"

    if [ -n "$(git -C "$APP_DIR" status --porcelain 2>/dev/null)" ]; then
        warnings+=("checkout has uncommitted changes; the build will be stamped -dirty")
    fi

    if ! git -C "$APP_DIR" describe --exact-match --tags HEAD >/dev/null 2>&1; then
        warnings+=("HEAD is not tagged; tag and push before deploying to stamp a release version")
    fi

    if [ ${#warnings[@]} -eq 0 ]; then
        return 0
    fi

    # Printed before the --yes check on purpose. --yes waives the prompt, not the
    # warnings: an unattended deploy is exactly the case where the log is the only
    # record that the build was dirty or untagged.
    for warning in "${warnings[@]}"; do
        echo "    WARNING: $warning"
    done

    if [ "$ASSUME_YES" -eq 1 ]; then
        return 0
    fi

    # A deploy with no terminal — a timer, a pipe — must not block on a prompt.
    # The warnings are already printed and the version stamp records the same
    # facts, so continuing is the safe default there.
    if [ ! -t 0 ]; then
        echo "    (non-interactive, continuing)"

        return 0
    fi

    read -r -p "    Continue? [y/N] " reply
    case "$reply" in
        [yY] | [yY][eE][sS]) return 0 ;;
        *)
            echo "==> Deployment aborted."
            exit 1
            ;;
    esac
}

# Reads the migration version the database is currently at. migrate.sh prints
# progress around the command, so the bare number is selected rather than assumed
# to be the last line.
#
# `|| true` is load-bearing: this only feeds the manifest, and under `pipefail`
# a `grep` that matches nothing would otherwise fail the assignment it is called
# from and abort the whole deploy — silently, since stderr is discarded.
db_version() {
    "$SCRIPTS/migrate.sh" db-version 2>/dev/null | grep -E "^[0-9]+$" | tail -n 1 || true
}

# Keeps a copy of every binary this deploy built, so a rollback is a file copy
# rather than a rebuild from a detached HEAD. The directory is owned by the
# deploy account and mode 700, so nothing here needs sudo.
archive_build() {
    local db_before="$1" db_after="$2" dest binary

    dest="$ARCHIVE_DIR/$VERSION"

    echo "==> Archiving build to $dest..."
    mkdir -p "$dest"

    for binary in "${ARCHIVED_BINARIES[@]}"; do
        cp "$APP_BIN_DIR/$binary" "$dest/$binary"
    done

    # Read by humans, not by rollback.sh — it asks the archived migrate binary
    # directly what schema it expects, which cannot drift from the binary itself.
    cat > "$dest/manifest" <<MANIFEST
version           $VERSION
commit            $(git -C "$APP_DIR" rev-parse --short HEAD 2>/dev/null || echo unknown)
deployed          $(date -u +%Y-%m-%dT%H:%M:%SZ)
schema_version    $("$dest/migrate" schema-version 2>/dev/null || echo unknown)
db_version_before ${db_before:-unknown}
db_version_after  ${db_after:-unknown}
MANIFEST

    chmod -R 700 "$dest"

    prune_archives
}

# Keeps the newest ARCHIVE_RETENTION archived versions, by modification time.
# Directory names carry a version rather than a date, so they do not sort
# chronologically and mtime is the only ordering that means anything.
#
# Parsing `ls` is safe here and nowhere else: these names are `git describe`
# output, which cannot contain a space or a newline.
prune_archives() {
    local old

    [ -d "$ARCHIVE_DIR" ] || return 0

    # shellcheck disable=SC2012
    while IFS= read -r old; do
        [ -n "$old" ] || continue
        echo "    pruning old archive $(basename "$old")"
        rm -rf "$old"
    done < <(ls -dt "$ARCHIVE_DIR"/*/ 2>/dev/null | tail -n +$((ARCHIVE_RETENTION + 1)))
}

main() {
    local db_before db_after

    if [ "$EUID" -eq 0 ]; then
        echo "ERROR: Do not run this script as root."
        exit 1
    fi

    for arg in "$@"; do
        case "$arg" in
            --yes | -y) ASSUME_YES=1 ;;
            *)
                echo "ERROR: unknown option '$arg'. Usage: deploy.sh [--yes|-y]"
                exit 1
                ;;
        esac
    done

    "$SCRIPTS/pull.sh"

    VERSION="$(git -C "$APP_DIR" describe --tags --always --dirty 2>/dev/null || echo unknown)"

    confirm_version

    echo "==> Building javascript..."
    "$SCRIPTS/build-js.sh"

    echo "==> Building binaries..."
    "$SCRIPTS/build.sh"

    db_before="$(db_version)"

    echo "==> Snapshotting the database before migrating..."
    "$SCRIPTS/migrate.sh" snapshot

    echo "==> Running database migrations..."
    "$SCRIPTS/migrate.sh" up

    db_after="$(db_version)"

    archive_build "$db_before" "$db_after"

    echo "==> Installing updated binary to $MAIN_APP_PATH..."
    sudo cp "$BUILT_BIN_APP_PATH" "$MAIN_APP_PATH"
    sudo chown root:root "$MAIN_APP_PATH"
    sudo chmod 755 "$MAIN_APP_PATH"

    echo "==> Restarting service..."
    sudo systemctl restart ninete.service

    echo "==> Deployment successful!"
}

# `exit` shares this line deliberately. Bash parses one command at a time and
# seeks back to just past it before executing, so returning from a `main` that
# rewrote this file would resume at a stale offset in the new bytes. An `exit`
# on its own line lives in the old bytes and is never read.
main "$@"; exit
