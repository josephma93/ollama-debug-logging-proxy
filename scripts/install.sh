#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "install.sh currently supports macOS only." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

BIN_DIR="${BIN_DIR:-$HOME/bin}"
BINARY_NAME="${BINARY_NAME:-ollama-logging-proxy}"
BINARY_PATH="${BINARY_PATH:-$BIN_DIR/$BINARY_NAME}"
CMD_PACKAGE="${CMD_PACKAGE:-./cmd/ollama-proxy}"

mkdir -p "$BIN_DIR"

echo "Building $BINARY_NAME -> $BINARY_PATH"
(
  cd "$REPO_ROOT"
  go build -o "$BINARY_PATH" "$CMD_PACKAGE"
)

echo "Installing LaunchAgents with $BINARY_PATH"
BINARY_PATH="$BINARY_PATH" "$SCRIPT_DIR/install-launchd.sh"
