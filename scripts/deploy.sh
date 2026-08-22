#!/bin/bash
set -euo pipefail

SCRIPTS="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd -P "$SCRIPTS/.." && pwd)"

BUILT_BIN_APP_PATH="/srv/ninete/bin/ninete"
MAIN_APP_PATH="/usr/local/bin/ninete"

# Set by --yes/-y to skip the confirmation prompt.
ASSUME_YES=0

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
    local version warnings=() reply

    version="$(git -C "$APP_DIR" describe --tags --always --dirty 2>/dev/null || echo unknown)"

    echo "==> Deploying version $version"

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

main() {
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

    confirm_version

    echo "==> Building javascript..."
    "$SCRIPTS/build-js.sh"

    echo "==> Building binaries..."
    "$SCRIPTS/build.sh"

    echo "==> Running database migrations..."
    "$SCRIPTS/migrate.sh" up

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
