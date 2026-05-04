#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "install.sh currently supports macOS only." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PROXY_LABEL="${PROXY_LABEL:-com.joseph.ollama-proxy}"
OLLAMA_LABEL="${OLLAMA_LABEL:-com.joseph.ollama-server}"

BIN_DIR="${BIN_DIR:-$HOME/bin}"
BINARY_NAME="${BINARY_NAME:-ollama-logging-proxy}"
BINARY_PATH="${BINARY_PATH:-$BIN_DIR/$BINARY_NAME}"
CMD_PACKAGE="${CMD_PACKAGE:-./cmd/ollama-proxy}"

LAUNCH_AGENTS_DIR="${LAUNCH_AGENTS_DIR:-$HOME/Library/LaunchAgents}"
PROXY_TEMPLATE="${PROXY_TEMPLATE:-$REPO_ROOT/launchd/com.joseph.ollama-proxy.plist}"
OLLAMA_TEMPLATE="${OLLAMA_TEMPLATE:-$REPO_ROOT/launchd/com.joseph.ollama-private.plist}"
PROXY_PLIST_PATH="${PROXY_PLIST_PATH:-$LAUNCH_AGENTS_DIR/${PROXY_LABEL}.plist}"
OLLAMA_PLIST_PATH="${OLLAMA_PLIST_PATH:-$LAUNCH_AGENTS_DIR/${OLLAMA_LABEL}.plist}"

OLLAMA_BIN="${OLLAMA_BIN:-/Applications/Ollama.app/Contents/Resources/ollama}"
OLLAMA_HOST="${OLLAMA_HOST:-127.0.0.1:11435}"

OLLAMA_PROXY_LISTEN="${OLLAMA_PROXY_LISTEN:-0.0.0.0:11434}"
OLLAMA_PROXY_TARGET="${OLLAMA_PROXY_TARGET:-http://127.0.0.1:11435}"
OLLAMA_PROXY_LOG_DIR="${OLLAMA_PROXY_LOG_DIR:-$HOME/Library/Logs/ollama-proxy}"
OLLAMA_PROXY_RETENTION_DAYS="${OLLAMA_PROXY_RETENTION_DAYS:-10}"
OLLAMA_PROXY_MAX_BODY_BYTES="${OLLAMA_PROXY_MAX_BODY_BYTES:-10485760}"

OLLAMA_LOG_DIR="${OLLAMA_LOG_DIR:-$HOME/Library/Logs/ollama}"

if [[ ! -f "$PROXY_TEMPLATE" ]]; then
  echo "Missing proxy template: $PROXY_TEMPLATE" >&2
  exit 1
fi
if [[ ! -f "$OLLAMA_TEMPLATE" ]]; then
  echo "Missing ollama template: $OLLAMA_TEMPLATE" >&2
  exit 1
fi
if [[ ! -x "$OLLAMA_BIN" ]]; then
  echo "Ollama binary not found at $OLLAMA_BIN" >&2
  exit 1
fi

mkdir -p "$BIN_DIR" "$LAUNCH_AGENTS_DIR" "$OLLAMA_PROXY_LOG_DIR" "$OLLAMA_LOG_DIR"

echo "Building $BINARY_NAME -> $BINARY_PATH"
go build -o "$BINARY_PATH" "$CMD_PACKAGE"

echo "Installing LaunchAgent plists into $LAUNCH_AGENTS_DIR"
cp "$PROXY_TEMPLATE" "$PROXY_PLIST_PATH"
cp "$OLLAMA_TEMPLATE" "$OLLAMA_PLIST_PATH"

/usr/libexec/PlistBuddy -c "Set :Label $PROXY_LABEL" "$PROXY_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :ProgramArguments:0 $BINARY_PATH" "$PROXY_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :EnvironmentVariables:OLLAMA_PROXY_LISTEN $OLLAMA_PROXY_LISTEN" "$PROXY_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :EnvironmentVariables:OLLAMA_PROXY_TARGET $OLLAMA_PROXY_TARGET" "$PROXY_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :EnvironmentVariables:OLLAMA_PROXY_LOG_DIR $OLLAMA_PROXY_LOG_DIR" "$PROXY_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :EnvironmentVariables:OLLAMA_PROXY_RETENTION_DAYS $OLLAMA_PROXY_RETENTION_DAYS" "$PROXY_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :EnvironmentVariables:OLLAMA_PROXY_MAX_BODY_BYTES $OLLAMA_PROXY_MAX_BODY_BYTES" "$PROXY_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :StandardOutPath $OLLAMA_PROXY_LOG_DIR/stdout.log" "$PROXY_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :StandardErrorPath $OLLAMA_PROXY_LOG_DIR/stderr.log" "$PROXY_PLIST_PATH"

/usr/libexec/PlistBuddy -c "Set :Label $OLLAMA_LABEL" "$OLLAMA_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :ProgramArguments:0 $OLLAMA_BIN" "$OLLAMA_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :EnvironmentVariables:OLLAMA_HOST $OLLAMA_HOST" "$OLLAMA_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :StandardOutPath $OLLAMA_LOG_DIR/stdout.log" "$OLLAMA_PLIST_PATH"
/usr/libexec/PlistBuddy -c "Set :StandardErrorPath $OLLAMA_LOG_DIR/stderr.log" "$OLLAMA_PLIST_PATH"

UID_DOMAIN="gui/$(id -u)"

echo "Reloading LaunchAgents"
launchctl bootout "$UID_DOMAIN/$PROXY_LABEL" >/dev/null 2>&1 || true
launchctl bootout "$UID_DOMAIN/$OLLAMA_LABEL" >/dev/null 2>&1 || true
launchctl bootstrap "$UID_DOMAIN" "$OLLAMA_PLIST_PATH"
launchctl bootstrap "$UID_DOMAIN" "$PROXY_PLIST_PATH"
launchctl enable "$UID_DOMAIN/$OLLAMA_LABEL"
launchctl enable "$UID_DOMAIN/$PROXY_LABEL"
launchctl kickstart -k "$UID_DOMAIN/$OLLAMA_LABEL"
launchctl kickstart -k "$UID_DOMAIN/$PROXY_LABEL"

echo "Install complete."
echo "Proxy health check: curl -fsS http://127.0.0.1:11434/__ollama_logging_proxy/health"
