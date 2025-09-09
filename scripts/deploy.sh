#!/usr/bin/env bash
set -Eeuo pipefail

# --------------------------------------------- #
HOST="${HOST:-hetzner}"

BINARY_NAME="ad9311app"

REMOTE_BINARY_PATH="/usr/local/bin/"
REMOTE_TEMP_PATH="/home/ad9311/tmp"

LOCAL_BINARY="build/ad9311app"
# --------------------------------------------- #

eval "$(ssh-agent -s)"
ssh-add -t 600 ~/.ssh/hetzner_ssh_key

printf "\033[33m▶ Running linters\033[0m\n"
make lint

printf "\n\033[33m▶ Running tests\033[0m\n"
make test

printf "\n\033[33m▶ Building final binary\033[0m\n"
make build-final
[[ -f "$LOCAL_BINARY" ]] || {
  echo "❌ Missing $LOCAL_BINARY"
  exit 1
}

printf "\n✅ \033[32mLocal binary ready for deployment!\033[0m\n"

printf "\n\033[33m▶ Copying binary to server\033[0m\n"
scp "$LOCAL_BINARY" "$HOST:$REMOTE_TEMP_PATH/$BINARY_NAME"

printf "\n✅ \033[32mFiles transfered successfully\033[0m\n"

printf "\n\033[33m▶ Running migrations and restarting the server\033[0m\n"
ssh "$HOST" "set -Eeuo pipefail; \
  sudo mv /home/ad9311/tmp/$BINARY_NAME $REMOTE_BINARY_PATH/$BINARY_NAME; \
  sudo /etc/ad9311app/migrate.sh
  sudo systemctl restart $BINARY_NAME.service; \
"

printf "\n✅ \033[32mDeployment successful!\033[0m\n"
