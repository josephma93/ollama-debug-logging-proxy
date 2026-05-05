#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "uninstall.sh currently supports macOS only." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

BIN_DIR="${BIN_DIR:-$HOME/bin}"
BINARY_NAME="${BINARY_NAME:-ollama-logging-proxy}"
BINARY_PATH="${BINARY_PATH:-$BIN_DIR/$BINARY_NAME}"

if [[ -z "${REMOVE_BINARY:-}" ]]; then
  REMOVE_BINARY="1"
  if command -v brew >/dev/null 2>&1; then
    BREW_PREFIX="$(brew --prefix)"
    if [[ "$BINARY_PATH" == "$BREW_PREFIX/bin/$BINARY_NAME" ]]; then
      REMOVE_BINARY="0"
    fi
  fi
fi

"$SCRIPT_DIR/uninstall-launchd.sh"

if [[ "$REMOVE_BINARY" == "1" ]]; then
  echo "Removing binary $BINARY_PATH"
  rm -f "$BINARY_PATH"
else
  echo "Leaving binary in place at $BINARY_PATH"
fi

echo "Uninstall complete."
