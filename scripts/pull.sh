#!/bin/bash
set -euo pipefail

# Resolved with `cd -P` so the script works when invoked through a symlink
# (the host reaches it as /srv/ninete/scripts, a link into the checkout).
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

main "$@"
