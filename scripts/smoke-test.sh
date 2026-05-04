#!/usr/bin/env bash
set -euo pipefail

PROXY_URL="${PROXY_URL:-http://127.0.0.1:11434}"
OLLAMA_PROXY_LOG_DIR="${OLLAMA_PROXY_LOG_DIR:-$HOME/Library/Logs/ollama-proxy}"
MODEL_NAME="${MODEL_NAME:-smoke-test-model}"

TMP_BODY="$(mktemp)"
TMP_RESPONSE="$(mktemp)"
trap 'rm -f "$TMP_BODY" "$TMP_RESPONSE"' EXIT

cat > "$TMP_BODY" <<EOF
{"model":"$MODEL_NAME","prompt":"proxy smoke test","stream":false}
EOF

echo "Checking proxy health endpoint..."
curl -fsS "$PROXY_URL/__ollama_logging_proxy/health" >/dev/null

echo "Sending tapped request to /api/generate..."
status_code="$(
  curl -sS \
    -o "$TMP_RESPONSE" \
    -w "%{http_code}" \
    -H "Content-Type: application/json" \
    -X POST "$PROXY_URL/api/generate" \
    --data @"$TMP_BODY" || true
)"

if [[ "$status_code" == "000" ]]; then
  echo "Request failed: could not connect to proxy/upstream." >&2
  exit 1
fi

today="$(date +%F)"
body_log="$OLLAMA_PROXY_LOG_DIR/body-$today.jsonl"

echo "Looking for log entry in $body_log..."
found=0
for _ in {1..15}; do
  if [[ -f "$body_log" ]] && tail -n 200 "$body_log" | grep -Fq '"path":"/api/generate"'; then
    found=1
    break
  fi
  sleep 1
done

if [[ "$found" -ne 1 ]]; then
  echo "Did not find expected /api/generate entry in $body_log" >&2
  echo "Last response body from proxy/upstream:" >&2
  cat "$TMP_RESPONSE" >&2
  exit 1
fi

echo "Smoke test passed. HTTP status: $status_code"
