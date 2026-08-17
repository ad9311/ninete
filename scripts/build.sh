#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd -P "$SCRIPT_DIR/.." && pwd)"

# Build output lives outside the checkout, mode 700.
APP_BIN_DIR="/srv/ninete/bin"

MAIN_BIN="$APP_BIN_DIR/ninete"
MIGRATION_BIN="$APP_BIN_DIR/migrate"
TASK_BIN="$APP_BIN_DIR/task"

# Build environment
GO_ENVS='CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc'
GO_FLAGS='-trimpath -ldflags=-s -ldflags=-w -buildvcs=false'

main() {
    if [ "$EUID" -eq 0 ]; then
        echo "ERROR: Do not run this script as root."
        exit 1
    fi

    echo "==> Switching to application directory..."
    cd "$APP_DIR"

    echo "==> Building migration binary..."
    # Word splitting on $GO_ENVS and $GO_FLAGS is intended.
    # shellcheck disable=SC2086
    env $GO_ENVS go build $GO_FLAGS -o "$MIGRATION_BIN" "$APP_DIR/cmd/migrate"

    echo "==> Building task binary..."
    # shellcheck disable=SC2086
    env $GO_ENVS go build $GO_FLAGS -o "$TASK_BIN" "$APP_DIR/cmd/task"

    echo "==> Building main application binary..."
    # shellcheck disable=SC2086
    env $GO_ENVS go build $GO_FLAGS -o "$MAIN_BIN" "$APP_DIR/cmd/ninete"

    echo "==> Securing binary directory permissions..."
    chmod -R 700 "$APP_BIN_DIR"

    echo "==> Build completed successfully."
}

main "$@"
