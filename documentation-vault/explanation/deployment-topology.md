# Deployment Topology

The intended deployment is a two-process topology:

1. `dev.ollama.server` runs Ollama privately on `127.0.0.1:11435`
2. `dev.ollama.logging-proxy` listens on `11434` and forwards to `http://127.0.0.1:11435`

That means the user-facing Ollama endpoint stays on `11434`, but the process serving `11434` changes from Ollama itself to this proxy.

Expected steady state after install:

- `dev.ollama.server` LaunchAgent runs `/Applications/Ollama.app/Contents/Resources/ollama serve`
- `OLLAMA_HOST=127.0.0.1:11435`
- `dev.ollama.logging-proxy` LaunchAgent runs `ollama-logging-proxy serve`
- proxy listens on `0.0.0.0:11434` by default
- proxy forwards upstream to `http://127.0.0.1:11435`
- logs are written to `~/Library/Logs/ollama-proxy/body-YYYY-MM-DD.jsonl`

For log retention behavior, see [[0009-filename-based-retention]].

## Related

- [[overview]]
- [[install]]
- [[validate-setup]]
- [[0001-reverse-proxy-not-fork|0001 — Reverse Proxy in Front of Ollama, Don't Fork or Patch]]
- [[0002-two-process-launchd-topology|0002 — Two-Process launchd Topology]]
- [[0011-proxy-owned-health-endpoint|0011 — Health Endpoint]]
- [[home]]
