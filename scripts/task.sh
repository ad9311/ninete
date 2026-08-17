#!/bin/bash
set -euo pipefail

ENV_FILE="/etc/ninete/env"
TASK_BIN="/srv/ninete/bin/task"

main() {
    if [ "$EUID" -eq 0 ]; then
        echo "ERROR: Do not run this script as root."
        exit 1
    fi

    echo "==> Loading environment variables..."
    set -o allexport
    # shellcheck disable=SC1090,SC1091
    source "$ENV_FILE"
    set +o allexport

    # If no arguments were provided, show help and exit gracefully
    if [[ $# -eq 0 ]]; then
        echo "==> No task command supplied. Showing help..."
        "$TASK_BIN" help
        exit 0
    fi

    echo "==> Running task: $*"
    "$TASK_BIN" "$@"

    echo "==> Task completed successfully."
}

# `exit` shares this line deliberately. Bash parses one command at a time and
# seeks back to just past it before executing, so returning from a `main` that
# rewrote this file would resume at a stale offset in the new bytes. An `exit`
# on its own line lives in the old bytes and is never read.
main "$@"; exit
