#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "install-launchd.sh currently supports macOS only." >&2
  exit 1
fi

plist_read_string() {
  local file="$1"
  local key="$2"

  if [[ ! -f "$file" ]]; then
    return 0
  fi

  /usr/libexec/PlistBuddy -c "Print $key" "$file" 2>/dev/null || true
}

plist_set_string() {
  local file="$1"
  local key="$2"
  local value="$3"

  if /usr/libexec/PlistBuddy -c "Print $key" "$file" >/dev/null 2>&1; then
    /usr/libexec/PlistBuddy -c "Set $key $value" "$file"
    return
  fi

  /usr/libexec/PlistBuddy -c "Add $key string $value" "$file"
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGE_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PROXY_LABEL="${PROXY_LABEL:-com.joseph.ollama-proxy}"
OLLAMA_LABEL="${OLLAMA_LABEL:-com.joseph.ollama-server}"
BINARY_NAME="${BINARY_NAME:-ollama-logging-proxy}"

if [[ -z "${BINARY_PATH:-}" ]]; then
  BINARY_PATH="$(command -v "$BINARY_NAME" || true)"
fi

if [[ -z "$BINARY_PATH" ]]; then
  echo "Could not resolve $BINARY_NAME on PATH. Set BINARY_PATH explicitly." >&2
  exit 1
fi
if [[ ! -x "$BINARY_PATH" ]]; then
  echo "Proxy binary not executable at $BINARY_PATH" >&2
  exit 1
fi

LAUNCH_AGENTS_DIR="${LAUNCH_AGENTS_DIR:-$HOME/Library/LaunchAgents}"
PROXY_TEMPLATE="${PROXY_TEMPLATE:-$PACKAGE_ROOT/launchd/com.joseph.ollama-proxy.plist}"
OLLAMA_TEMPLATE="${OLLAMA_TEMPLATE:-$PACKAGE_ROOT/launchd/com.joseph.ollama-private.plist}"
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

EXISTING_OLLAMA_DEBUG="$(plist_read_string "$OLLAMA_PLIST_PATH" ":EnvironmentVariables:OLLAMA_DEBUG")"
EXISTING_OLLAMA_MODELS="$(plist_read_string "$OLLAMA_PLIST_PATH" ":EnvironmentVariables:OLLAMA_MODELS")"
EXISTING_OLLAMA_HOME="$(plist_read_string "$OLLAMA_PLIST_PATH" ":EnvironmentVariables:HOME")"
EXISTING_OLLAMA_PATH="$(plist_read_string "$OLLAMA_PLIST_PATH" ":EnvironmentVariables:PATH")"

OLLAMA_DEBUG="${OLLAMA_DEBUG:-$EXISTING_OLLAMA_DEBUG}"
OLLAMA_MODELS="${OLLAMA_MODELS:-$EXISTING_OLLAMA_MODELS}"
LAUNCHD_HOME="${LAUNCHD_HOME:-${EXISTING_OLLAMA_HOME:-$HOME}}"
LAUNCHD_PATH="${LAUNCHD_PATH:-${EXISTING_OLLAMA_PATH:-/usr/bin:/bin:/usr/sbin:/sbin}}"

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

mkdir -p "$LAUNCH_AGENTS_DIR" "$OLLAMA_PROXY_LOG_DIR" "$OLLAMA_LOG_DIR"

echo "Installing LaunchAgent plists into $LAUNCH_AGENTS_DIR"
cp "$PROXY_TEMPLATE" "$PROXY_PLIST_PATH"
cp "$OLLAMA_TEMPLATE" "$OLLAMA_PLIST_PATH"

plist_set_string "$PROXY_PLIST_PATH" ":Label" "$PROXY_LABEL"
plist_set_string "$PROXY_PLIST_PATH" ":ProgramArguments:0" "$BINARY_PATH"
plist_set_string "$PROXY_PLIST_PATH" ":EnvironmentVariables:OLLAMA_PROXY_LISTEN" "$OLLAMA_PROXY_LISTEN"
plist_set_string "$PROXY_PLIST_PATH" ":EnvironmentVariables:OLLAMA_PROXY_TARGET" "$OLLAMA_PROXY_TARGET"
plist_set_string "$PROXY_PLIST_PATH" ":EnvironmentVariables:OLLAMA_PROXY_LOG_DIR" "$OLLAMA_PROXY_LOG_DIR"
plist_set_string "$PROXY_PLIST_PATH" ":EnvironmentVariables:OLLAMA_PROXY_RETENTION_DAYS" "$OLLAMA_PROXY_RETENTION_DAYS"
plist_set_string "$PROXY_PLIST_PATH" ":EnvironmentVariables:OLLAMA_PROXY_MAX_BODY_BYTES" "$OLLAMA_PROXY_MAX_BODY_BYTES"
plist_set_string "$PROXY_PLIST_PATH" ":StandardOutPath" "$OLLAMA_PROXY_LOG_DIR/stdout.log"
plist_set_string "$PROXY_PLIST_PATH" ":StandardErrorPath" "$OLLAMA_PROXY_LOG_DIR/stderr.log"

plist_set_string "$OLLAMA_PLIST_PATH" ":Label" "$OLLAMA_LABEL"
plist_set_string "$OLLAMA_PLIST_PATH" ":ProgramArguments:0" "$OLLAMA_BIN"
plist_set_string "$OLLAMA_PLIST_PATH" ":EnvironmentVariables:OLLAMA_HOST" "$OLLAMA_HOST"
plist_set_string "$OLLAMA_PLIST_PATH" ":EnvironmentVariables:HOME" "$LAUNCHD_HOME"
plist_set_string "$OLLAMA_PLIST_PATH" ":EnvironmentVariables:PATH" "$LAUNCHD_PATH"
plist_set_string "$OLLAMA_PLIST_PATH" ":StandardOutPath" "$OLLAMA_LOG_DIR/stdout.log"
plist_set_string "$OLLAMA_PLIST_PATH" ":StandardErrorPath" "$OLLAMA_LOG_DIR/stderr.log"

if [[ -n "$OLLAMA_DEBUG" ]]; then
  plist_set_string "$OLLAMA_PLIST_PATH" ":EnvironmentVariables:OLLAMA_DEBUG" "$OLLAMA_DEBUG"
fi

if [[ -n "$OLLAMA_MODELS" ]]; then
  plist_set_string "$OLLAMA_PLIST_PATH" ":EnvironmentVariables:OLLAMA_MODELS" "$OLLAMA_MODELS"
fi

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
