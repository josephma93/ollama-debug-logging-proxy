#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "uninstall-launchd.sh currently supports macOS only." >&2
  exit 1
fi

PROXY_LABEL="${PROXY_LABEL:-dev.ollama.logging-proxy}"
OLLAMA_LABEL="${OLLAMA_LABEL:-dev.ollama.server}"

LAUNCH_AGENTS_DIR="${LAUNCH_AGENTS_DIR:-$HOME/Library/LaunchAgents}"
PROXY_PLIST_PATH="${PROXY_PLIST_PATH:-$LAUNCH_AGENTS_DIR/${PROXY_LABEL}.plist}"
OLLAMA_PLIST_PATH="${OLLAMA_PLIST_PATH:-$LAUNCH_AGENTS_DIR/${OLLAMA_LABEL}.plist}"

REMOVE_LOGS="${REMOVE_LOGS:-0}"
OLLAMA_PROXY_LOG_DIR="${OLLAMA_PROXY_LOG_DIR:-$HOME/Library/Logs/ollama-proxy}"

UID_DOMAIN="gui/$(id -u)"

echo "Stopping LaunchAgents"
launchctl bootout "$UID_DOMAIN/$PROXY_LABEL" >/dev/null 2>&1 || true
launchctl bootout "$UID_DOMAIN/$OLLAMA_LABEL" >/dev/null 2>&1 || true
launchctl disable "$UID_DOMAIN/$PROXY_LABEL" >/dev/null 2>&1 || true
launchctl disable "$UID_DOMAIN/$OLLAMA_LABEL" >/dev/null 2>&1 || true

echo "Removing LaunchAgent files"
rm -f "$PROXY_PLIST_PATH" "$OLLAMA_PLIST_PATH"

if [[ "$REMOVE_LOGS" == "1" ]]; then
  echo "Removing proxy logs in $OLLAMA_PROXY_LOG_DIR"
  rm -rf "$OLLAMA_PROXY_LOG_DIR"
fi

echo "LaunchAgent uninstall complete."
