# Validate the Setup

After `ollama-logging-proxy-install`, these checks should succeed:

```bash
curl -fsS http://127.0.0.1:11434/__ollama_logging_proxy/health
curl -fsS http://127.0.0.1:11434/api/version
launchctl print gui/$(id -u)/dev.ollama.logging-proxy | rg 'program ='
launchctl print gui/$(id -u)/dev.ollama.server | rg 'program ='
```

Expected shape:

- proxy health endpoint returns `{"ok":true,"service":"ollama-logging-proxy"}`
- `/api/version` succeeds through the proxy
- proxy LaunchAgent points at the installed `ollama-logging-proxy` binary
- Ollama LaunchAgent is loaded separately and bound to `127.0.0.1:11435`

To inspect captured traffic with the built-in CLI:

```bash
ollama-logging-proxy tail 20
```

## Related

- [[install]]
- [[cli]]
- [[0011-proxy-owned-health-endpoint|0011 — Health Endpoint]]
- [[home]]
