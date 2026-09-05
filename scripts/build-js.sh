#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd -P "$SCRIPT_DIR/.." && pwd)"

main() {
    if [ "$EUID" -eq 0 ]; then
        echo "ERROR: Do not run this script as root."
        exit 1
    fi

    echo "==> Switching to application directory..."
    cd "$APP_DIR"

    echo "==> Installing dependencies..."
    bun install

    echo "==> Building the frontend bundle and stylesheet..."
    make build-static

    echo "==> Frontend assets successfully built."
}

# `exit` shares this line deliberately. Bash parses one command at a time and
# seeks back to just past it before executing, so returning from a `main` that
# rewrote this file would resume at a stale offset in the new bytes. An `exit`
# on its own line lives in the old bytes and is never read.
main "$@"; exit
