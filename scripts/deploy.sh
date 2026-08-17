#!/bin/bash
set -euo pipefail

SCRIPTS="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

BUILT_BIN_APP_PATH="/srv/ninete/bin/ninete"
MAIN_APP_PATH="/usr/local/bin/ninete"

main() {
    if [ "$EUID" -eq 0 ]; then
        echo "ERROR: Do not run this script as root."
        exit 1
    fi

    "$SCRIPTS/pull.sh"

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

main "$@"
