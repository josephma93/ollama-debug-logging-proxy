# launchd Templates

Templates live in `launchd/`:

- `dev.ollama.logging-proxy.plist`
- `dev.ollama.server.plist`

Default values:

- proxy listener: `0.0.0.0:11434`
- private Ollama upstream: `127.0.0.1:11435`
- proxy log dir: `~/Library/Logs/ollama-proxy`
- retention: `10` days

## Related

- [[0002-two-process-launchd-topology|0002 — Two-Process launchd Topology]]
- [[0003-user-launchagent-scope|0003 — macOS User LaunchAgent]]
- [[0016-launchagent-labels-dev-ollama|0016 — Rename LaunchAgent Labels]]
- [[install]]
- [[home]]
