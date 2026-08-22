#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd -P "$SCRIPT_DIR/.." && pwd)"

# Build output lives outside the checkout, mode 700.
APP_BIN_DIR="/srv/ninete/bin"

MAIN_BIN="$APP_BIN_DIR/ninete"
MIGRATION_BIN="$APP_BIN_DIR/migrate"
TASK_BIN="$APP_BIN_DIR/task"

# Build identity, linked into internal/prog with -X and reported by the boot log
# and the `version` command. Derived from git, so it cannot drift from what was
# actually built. Each falls back to a literal: a deploy must not die because the
# checkout is unreadable by git, and "unknown" is a better outcome than no binary.
# -buildvcs=false stays below, so this is the only source of build identity.
VERSION_PKG="github.com/ad9311/ninete/internal/prog"
VERSION="$(git -C "$APP_DIR" describe --tags --always --dirty 2>/dev/null || echo unknown)"
COMMIT="$(git -C "$APP_DIR" rev-parse --short HEAD 2>/dev/null || echo unknown)"
BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# -s -w strip the symbol table and DWARF; the -X values ride in the same argument
# because -ldflags takes exactly one.
LD_FLAGS="-s -w"
LD_FLAGS="$LD_FLAGS -X ${VERSION_PKG}.Version=${VERSION}"
LD_FLAGS="$LD_FLAGS -X ${VERSION_PKG}.Commit=${COMMIT}"
LD_FLAGS="$LD_FLAGS -X ${VERSION_PKG}.BuildTime=${BUILD_TIME}"

# Build environment. Arrays, not strings: -ldflags takes a single quoted argument
# and a string expanded unquoted cannot carry one.
GO_ENVS=(CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc)
GO_FLAGS=(-trimpath "-ldflags=$LD_FLAGS" -buildvcs=false)

main() {
    if [ "$EUID" -eq 0 ]; then
        echo "ERROR: Do not run this script as root."
        exit 1
    fi

    echo "==> Building version $VERSION ($COMMIT)..."

    echo "==> Switching to application directory..."
    cd "$APP_DIR"

    echo "==> Building migration binary..."
    env "${GO_ENVS[@]}" go build "${GO_FLAGS[@]}" -o "$MIGRATION_BIN" "$APP_DIR/cmd/migrate"

    echo "==> Building task binary..."
    env "${GO_ENVS[@]}" go build "${GO_FLAGS[@]}" -o "$TASK_BIN" "$APP_DIR/cmd/task"

    echo "==> Building main application binary..."
    env "${GO_ENVS[@]}" go build "${GO_FLAGS[@]}" -o "$MAIN_BIN" "$APP_DIR/cmd/ninete"

    echo "==> Securing binary directory permissions..."
    chmod -R 700 "$APP_BIN_DIR"

    echo "==> Build completed successfully."
}

# `exit` shares this line deliberately. Bash parses one command at a time and
# seeks back to just past it before executing, so returning from a `main` that
# rewrote this file would resume at a stale offset in the new bytes. An `exit`
# on its own line lives in the old bytes and is never read.
main "$@"; exit
