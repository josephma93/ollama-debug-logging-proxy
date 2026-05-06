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

log() {
  echo "[install-launchd] $*"
}

fail() {
  local line="${1:-unknown}"
  shift || true
  trap - ERR
  echo "[install-launchd] ERROR at line $line: $*" >&2
  print_recovery_commands
  exit 1
}

fail_without_recovery() {
  echo "[install-launchd] ERROR: $*" >&2
  exit 1
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGE_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PROXY_LABEL="${PROXY_LABEL:-dev.ollama.logging-proxy}"
OLLAMA_LABEL="${OLLAMA_LABEL:-dev.ollama.server}"
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
PROXY_TEMPLATE="${PROXY_TEMPLATE:-$PACKAGE_ROOT/launchd/dev.ollama.logging-proxy.plist}"
OLLAMA_TEMPLATE="${OLLAMA_TEMPLATE:-$PACKAGE_ROOT/launchd/dev.ollama.server.plist}"
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
OLLAMA_HEALTH_HOST="${OLLAMA_HEALTH_HOST:-127.0.0.1}"
OLLAMA_HEALTH_PORT="${OLLAMA_HEALTH_PORT:-11435}"
PROXY_HEALTH_URL="${PROXY_HEALTH_URL:-http://127.0.0.1:11434/__ollama_logging_proxy/health}"
BOOTSTRAP_RETRY_ATTEMPTS="${BOOTSTRAP_RETRY_ATTEMPTS:-3}"
BOOTSTRAP_RETRY_BASE_DELAY="${BOOTSTRAP_RETRY_BASE_DELAY:-1}"
HEALTH_RETRY_ATTEMPTS="${HEALTH_RETRY_ATTEMPTS:-5}"
HEALTH_RETRY_DELAY="${HEALTH_RETRY_DELAY:-1}"

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

UID_DOMAIN="gui/$(id -u)"
TEMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/ollama-proxy-install.XXXXXX")"
PROXY_STAGE_PLIST="$TEMP_DIR/${PROXY_LABEL}.plist"
OLLAMA_STAGE_PLIST="$TEMP_DIR/${OLLAMA_LABEL}.plist"
trap 'rm -rf "$TEMP_DIR"' EXIT

print_recovery_commands() {
  cat >&2 <<EOF
[install-launchd] Recovery commands:
launchctl bootstrap "$UID_DOMAIN" "$OLLAMA_PLIST_PATH"
launchctl enable "$UID_DOMAIN/$OLLAMA_LABEL"
launchctl kickstart -k "$UID_DOMAIN/$OLLAMA_LABEL"
launchctl bootstrap "$UID_DOMAIN" "$PROXY_PLIST_PATH"
launchctl enable "$UID_DOMAIN/$PROXY_LABEL"
launchctl kickstart -k "$UID_DOMAIN/$PROXY_LABEL"
EOF
}

on_error() {
  local line="${1:-unknown}"
  local exit_code="${2:-1}"
  echo "[install-launchd] ERROR: install failed near line $line (exit $exit_code)" >&2
  print_recovery_commands
  exit "$exit_code"
}

trap 'on_error "$LINENO" "$?"' ERR

normalize_plist() {
  local file="$1"
  plutil -convert xml1 -o - "$file"
}

plists_equal() {
  local left="$1"
  local right="$2"

  if [[ ! -f "$left" || ! -f "$right" ]]; then
    return 1
  fi

  cmp -s <(normalize_plist "$left") <(normalize_plist "$right")
}

lint_plist() {
  local file="$1"

  if ! plutil -lint "$file" >/dev/null; then
    fail "$LINENO" "invalid plist: $file"
  fi
}

service_loaded() {
  local label="$1"
  launchctl print "$UID_DOMAIN/$label" >/dev/null 2>&1
}

wait_for_service_unload() {
  local label="$1"
  local attempts="${2:-20}"
  local delay="${3:-0.25}"
  local i

  for ((i = 1; i <= attempts; i++)); do
    if ! service_loaded "$label"; then
      return 0
    fi
    sleep "$delay"
  done

  return 1
}

verify_ollama_health() {
  local attempts="${1:-$HEALTH_RETRY_ATTEMPTS}"
  local delay="${2:-$HEALTH_RETRY_DELAY}"
  local i

  for ((i = 1; i <= attempts; i++)); do
    if nc -z -G 1 "$OLLAMA_HEALTH_HOST" "$OLLAMA_HEALTH_PORT" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$delay"
  done

  return 1
}

verify_proxy_health() {
  local attempts="${1:-$HEALTH_RETRY_ATTEMPTS}"
  local delay="${2:-$HEALTH_RETRY_DELAY}"
  local i

  for ((i = 1; i <= attempts; i++)); do
    if curl --max-time 2 -fsS "$PROXY_HEALTH_URL" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$delay"
  done

  return 1
}

should_retry_bootstrap() {
  local stderr_text="$1"
  local had_bootout="$2"

  [[ "$had_bootout" == "1" ]] || return 1

  if [[ "$stderr_text" == *"Input/output error"* || "$stderr_text" == *"Bootstrap failed: 5"* ]]; then
    return 0
  fi

  return 1
}

bootstrap_service() {
  local label="$1"
  local plist_path="$2"
  local had_bootout="$3"
  local attempts="${4:-$BOOTSTRAP_RETRY_ATTEMPTS}"
  local delay="${5:-$BOOTSTRAP_RETRY_BASE_DELAY}"
  local try
  local output=""

  for ((try = 1; try <= attempts; try++)); do
    output="$(launchctl bootstrap "$UID_DOMAIN" "$plist_path" 2>&1)" && return 0

    if should_retry_bootstrap "$output" "$had_bootout" && (( try < attempts )); then
      log "$label bootstrap hit transient launchd error; retrying ($try/$attempts)"
      sleep "$delay"
      delay=$((delay * 2))
      continue
    fi

    echo "$output" >&2
    fail "$LINENO" "$label bootstrap failed"
  done
}

install_plist_if_changed() {
  local stage_path="$1"
  local target_path="$2"
  local changed="$3"

  if [[ "$changed" == "1" ]]; then
    cp "$stage_path" "$target_path"
  fi
}

reconcile_service() {
  local label="$1"
  local target_plist="$2"
  local stage_plist="$3"
  local changed="$4"
  local verify_fn="$5"
  local health_desc="$6"
  local loaded="0"
  local had_bootout="0"
  local bootout_output=""
  local bootout_status=0

  if service_loaded "$label"; then
    loaded="1"
  fi

  if [[ "$changed" == "0" && "$loaded" == "1" ]] && "$verify_fn"; then
    log "$label plist unchanged; $health_desc healthy; skipping restart"
    return 0
  fi

  if [[ "$changed" == "1" ]]; then
    log "$label plist changed; installing updated plist and reconciling"
  elif [[ "$loaded" == "0" ]]; then
    log "$label plist unchanged; service not loaded; reconciling"
  else
    log "$label plist unchanged; $health_desc unhealthy; reconciling"
  fi

  install_plist_if_changed "$stage_plist" "$target_plist" "$changed"

  if service_loaded "$label"; then
    log "$label bootout"
    bootout_output="$(launchctl bootout "$UID_DOMAIN/$label" 2>&1)" || bootout_status=$?
    if (( bootout_status == 0 )); then
      had_bootout="1"
    else
      log "$label bootout returned non-zero; continuing to state check"
    fi
  else
    log "$label not currently loaded; skipping bootout"
  fi

  if ! wait_for_service_unload "$label"; then
    if (( bootout_status != 0 )); then
      if [[ -n "$bootout_output" ]]; then
        echo "$bootout_output" >&2
      fi
      fail "$LINENO" "$label failed to unload after bootout error"
    fi
    fail "$LINENO" "$label did not fully unload after bootout"
  fi

  log "$label bootstrap"
  bootstrap_service "$label" "$target_plist" "$had_bootout"
  log "$label enable"
  launchctl enable "$UID_DOMAIN/$label"
  log "$label kickstart"
  launchctl kickstart -k "$UID_DOMAIN/$label"

  if ! "$verify_fn"; then
    fail "$LINENO" "$label failed verification ($health_desc)"
  fi

  log "$label verified healthy"
}

mkdir -p "$LAUNCH_AGENTS_DIR" "$OLLAMA_PROXY_LOG_DIR" "$OLLAMA_LOG_DIR"

log "Rendering staged LaunchAgent plists in $TEMP_DIR"
cp "$PROXY_TEMPLATE" "$PROXY_STAGE_PLIST"
cp "$OLLAMA_TEMPLATE" "$OLLAMA_STAGE_PLIST"

plist_set_string "$PROXY_STAGE_PLIST" ":Label" "$PROXY_LABEL"
plist_set_string "$PROXY_STAGE_PLIST" ":ProgramArguments:0" "$BINARY_PATH"
plist_set_string "$PROXY_STAGE_PLIST" ":EnvironmentVariables:OLLAMA_PROXY_LISTEN" "$OLLAMA_PROXY_LISTEN"
plist_set_string "$PROXY_STAGE_PLIST" ":EnvironmentVariables:OLLAMA_PROXY_TARGET" "$OLLAMA_PROXY_TARGET"
plist_set_string "$PROXY_STAGE_PLIST" ":EnvironmentVariables:OLLAMA_PROXY_LOG_DIR" "$OLLAMA_PROXY_LOG_DIR"
plist_set_string "$PROXY_STAGE_PLIST" ":EnvironmentVariables:OLLAMA_PROXY_RETENTION_DAYS" "$OLLAMA_PROXY_RETENTION_DAYS"
plist_set_string "$PROXY_STAGE_PLIST" ":EnvironmentVariables:OLLAMA_PROXY_MAX_BODY_BYTES" "$OLLAMA_PROXY_MAX_BODY_BYTES"
plist_set_string "$PROXY_STAGE_PLIST" ":StandardOutPath" "$OLLAMA_PROXY_LOG_DIR/stdout.log"
plist_set_string "$PROXY_STAGE_PLIST" ":StandardErrorPath" "$OLLAMA_PROXY_LOG_DIR/stderr.log"

plist_set_string "$OLLAMA_STAGE_PLIST" ":Label" "$OLLAMA_LABEL"
plist_set_string "$OLLAMA_STAGE_PLIST" ":ProgramArguments:0" "$OLLAMA_BIN"
plist_set_string "$OLLAMA_STAGE_PLIST" ":EnvironmentVariables:OLLAMA_HOST" "$OLLAMA_HOST"
plist_set_string "$OLLAMA_STAGE_PLIST" ":EnvironmentVariables:HOME" "$LAUNCHD_HOME"
plist_set_string "$OLLAMA_STAGE_PLIST" ":EnvironmentVariables:PATH" "$LAUNCHD_PATH"
plist_set_string "$OLLAMA_STAGE_PLIST" ":StandardOutPath" "$OLLAMA_LOG_DIR/stdout.log"
plist_set_string "$OLLAMA_STAGE_PLIST" ":StandardErrorPath" "$OLLAMA_LOG_DIR/stderr.log"

if [[ -n "$OLLAMA_DEBUG" ]]; then
  plist_set_string "$OLLAMA_STAGE_PLIST" ":EnvironmentVariables:OLLAMA_DEBUG" "$OLLAMA_DEBUG"
fi

if [[ -n "$OLLAMA_MODELS" ]]; then
  plist_set_string "$OLLAMA_STAGE_PLIST" ":EnvironmentVariables:OLLAMA_MODELS" "$OLLAMA_MODELS"
fi

lint_plist "$OLLAMA_STAGE_PLIST"
lint_plist "$PROXY_STAGE_PLIST"

PROXY_CHANGED="1"
OLLAMA_CHANGED="1"

if plists_equal "$PROXY_STAGE_PLIST" "$PROXY_PLIST_PATH"; then
  PROXY_CHANGED="0"
fi

if plists_equal "$OLLAMA_STAGE_PLIST" "$OLLAMA_PLIST_PATH"; then
  OLLAMA_CHANGED="0"
fi

log "Reconciling LaunchAgents in dependency order"
reconcile_service "$OLLAMA_LABEL" "$OLLAMA_PLIST_PATH" "$OLLAMA_STAGE_PLIST" "$OLLAMA_CHANGED" verify_ollama_health "tcp probe ${OLLAMA_HEALTH_HOST}:${OLLAMA_HEALTH_PORT}"
reconcile_service "$PROXY_LABEL" "$PROXY_PLIST_PATH" "$PROXY_STAGE_PLIST" "$PROXY_CHANGED" verify_proxy_health "proxy health endpoint"

log "Install complete."
log "Proxy health check: curl -fsS $PROXY_HEALTH_URL"
