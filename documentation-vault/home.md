# Ollama Logging Proxy

> TODO: split into explanation/, reference/, how-to/ per vault-conventions. The sections below mix system overview, install runbook, CLI reference, and release model — each belongs in its own Diátaxis bucket. See [[vault-conventions]] for placement rules.

## Library

This project library is organized as a small linked set of notes:

- [[prd|PRD]]
- [[execution-plan|Execution Plan]]
- [[vault-conventions]]
- [[glossary]]

Use this note as the operator and system overview. Use the PRD for product requirements and the execution plan for engineering sequencing.

## Purpose

`ollama-logging-proxy` is a reverse proxy that sits in front of Ollama, preserves normal API behavior, and writes tapped request and response bodies to daily JSONL logs.

This project is designed primarily for macOS user-level `launchd` deployment.

## Deployment Model

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

## Installer Contract

`ollama-logging-proxy-install` is the command that rewires the local Ollama setup. It:

- installs or updates two user LaunchAgents in `~/Library/LaunchAgents`
- changes Ollama from public `11434` serving to private `127.0.0.1:11435`
- starts the proxy in front of Ollama on `11434`
- preserves existing Ollama `OLLAMA_DEBUG`, `OLLAMA_MODELS`, `HOME`, and `PATH` values across re-runs unless explicitly overridden

This is intentionally a user-scope `launchd` setup, not a system daemon install.

What it does not do:

- it does not use `brew services`
- it does not install anything into `/Library/LaunchDaemons`
- it does not keep Ollama directly exposed on `11434`
- it does not change non-macOS service managers
- it does not promise cross-platform deployment behavior

## Validate The Setup

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

## Runtime Details

### CLI

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

### launchd Templates

Templates live in `launchd/`:

- `dev.ollama.logging-proxy.plist`
- `dev.ollama.server.plist`

Default values:

- proxy listener: `0.0.0.0:11434`
- private Ollama upstream: `127.0.0.1:11435`
- proxy log dir: `~/Library/Logs/ollama-proxy`
- retention: `10` days

### Retention

The retention cleaner deletes only files matching `body-YYYY-MM-DD.jsonl`, based on the date encoded in the filename rather than file modification time.

### Scripts

Operational scripts live in `scripts/`:

- `install.sh`: source-checkout convenience path that builds the binary locally and then wires launchd
- `install-launchd.sh`: launchd setup path for an already-installed binary; intended to be idempotent and converge state
- `smoke-test.sh`: validates proxy health and basic tapped logging
- `uninstall-launchd.sh`: removes LaunchAgents and optionally proxy logs
- `uninstall.sh`: source-checkout convenience uninstall path

## Release Model

Homebrew release automation follows these rules:

- stable tags look like `v0.1.0`
- prerelease tags look like `v0.1.1-canary.1`, `v0.1.1-rc.1`, `v0.1.1-beta.1`, or `v0.1.1-alpha.1`
- any hyphen suffix after the numeric core marks a prerelease

Stable tags update the Homebrew formula. Prerelease tags publish release artifacts but do not update the stable Homebrew formula.

## Development Checks

Local development expects:

- Go `1.22`
- `just`
- `golangci-lint`
- `shellcheck`

Primary commands:

```bash
just hooks
just --list
just check
just test
```

## Related

- [[prd|PRD]]
- [[execution-plan|Execution Plan]]
- [[vault-conventions]]
- [[glossary]]
