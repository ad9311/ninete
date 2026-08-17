#!/bin/bash
set -euo pipefail

# Resolved with `cd -P` so the script works when invoked through a symlink
# (the host reaches it as /srv/ninete/scripts, a link into the checkout).
#
# This script rewrites itself: the `git pull` below can replace the very bytes
# bash is reading. See the trailing `main "$@"; exit` for why the `exit` has to
# share that line.
SCRIPT_DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd -P "$SCRIPT_DIR/.." && pwd)"

main() {
    if [ "$EUID" -eq 0 ]; then
        echo "ERROR: Do not run this script as root."
        exit 1
    fi

    echo "==> Switching to application directory..."
    cd "$APP_DIR"

    echo "==> Pulling latest changes from repository..."
    git pull --ff-only

    echo "==> Changes pulled"
}

# `exit` shares this line deliberately. Bash parses one command at a time and
# seeks back to just past it before executing, so returning from a `main` that
# rewrote this file would resume at a stale offset in the new bytes. An `exit`
# on its own line lives in the old bytes and is never read.
main "$@"; exit
