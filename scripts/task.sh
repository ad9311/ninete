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

main "$@"
