# CLI

The binary supports:

```bash
ollama-logging-proxy serve
ollama-logging-proxy health
ollama-logging-proxy tail [lines]
ollama-logging-proxy purge
```

- `serve`: starts the reverse proxy
- `health`: checks `GET /__ollama_logging_proxy/health` on the configured listener
- `tail`: prints recent lines from today's `body-YYYY-MM-DD.jsonl` file
- `purge`: runs one retention cleanup pass immediately

## Related

- [[0005-single-binary-subcommands|0005 — Single Binary Exposing Subcommands]]
- [[0011-proxy-owned-health-endpoint|0011 — Health Endpoint]]
- [[validate-setup]]
- [[home]]
