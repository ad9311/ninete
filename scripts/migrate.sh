#!/bin/bash
set -euo pipefail

ENV_FILE="/etc/ninete/env"
MIGRATION_BIN="/srv/ninete/bin/migrate"

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
    if [[ $# -lt 1 ]]; then
        echo "==> No migration command supplied. Showing help..."
        "$MIGRATION_BIN" help
        exit 0
    fi

    echo "==> Running migrations..."
    "$MIGRATION_BIN" "$@"

    echo "==> Migrations completed successfully."
}

main "$@"
