# Ollama Logging Proxy

Reverse proxy in front of Ollama that preserves normal API behavior and writes tapped request/response bodies to daily JSONL logs.

## Local quality checks

Install [`just`](https://github.com/casey/just) and `golangci-lint`, then run:

```bash
just --list
just check
```

`just check` runs: `fmt`, `vet`, `lint`, `test`, and `race`.

## Retention behavior

`internal/retention` deletes only files matching `body-YYYY-MM-DD.jsonl`, based on the date encoded in the filename (not mtime). Non-matching files are ignored.

Use the cleaner at startup and periodic cadence:

```go
cleaner := retention.NewCleaner(logDir, retentionDays)
errCh := cleaner.Start(ctx, time.Hour) // immediate run + hourly cleanup
```

## launchd templates

Templates are provided in [`launchd/`](launchd):

- `com.joseph.ollama-proxy.plist`
- `com.joseph.ollama-private.plist` (label: `com.joseph.ollama-server`)

Defaults follow the PRD:

- Proxy listener: `0.0.0.0:11434`
- Private Ollama upstream: `127.0.0.1:11435`
- Proxy log dir: `~/Library/Logs/ollama-proxy`
- Retention: `10` days

## macOS scripts

Scripts are in [`scripts/`](scripts):

```bash
./scripts/install.sh
./scripts/smoke-test.sh
./scripts/uninstall.sh
```

`install.sh` builds the proxy binary, installs launch agents into `~/Library/LaunchAgents`, configures PRD env vars, and bootstraps both agents.

`smoke-test.sh` checks `GET /__ollama_logging_proxy/health`, sends `POST /api/generate`, and verifies today’s `body-YYYY-MM-DD.jsonl` contains a `/api/generate` entry.

`uninstall.sh` unloads agents and removes installed files (binary removal on by default).

## CI

Pull requests run [`.github/workflows/ci.yml`](.github/workflows/ci.yml), which executes `just check`.
