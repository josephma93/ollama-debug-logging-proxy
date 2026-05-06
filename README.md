# Ollama Logging Proxy

Reverse proxy in front of Ollama that preserves normal API behavior and writes tapped request/response bodies to daily JSONL logs.

Detailed project and operator documentation lives in [home](documentation-vault/home.md).

## CLI

The binary supports:

```bash
ollama-logging-proxy serve
ollama-logging-proxy health
ollama-logging-proxy tail [lines]
ollama-logging-proxy purge
```

- `serve`: starts the reverse proxy.
- `health`: checks `GET /__ollama_logging_proxy/health` on the configured listener.
- `tail`: prints recent lines from today's `body-YYYY-MM-DD.jsonl` file (default 100).
- `purge`: runs one retention cleanup pass immediately.

## Homebrew

Supported install path:

```bash
brew tap josephma93/ollama-debug-logging-proxy https://github.com/josephma93/ollama-debug-logging-proxy
brew install josephma93/ollama-debug-logging-proxy/ollama-logging-proxy
```

This is the supported user path for stable releases. Modern Homebrew requires formulae to come from a tap; local `--HEAD` installs from `./Formula/...` are not supported here.

Homebrew only installs the binary and helper assets. It does not automatically rewrite your launchd topology.

To wire the proxy into launchd after `brew install`:

```bash
ollama-logging-proxy-install
```

Re-running `ollama-logging-proxy-install` is expected to be safe. The installer is designed to converge the current user LaunchAgent state to the desired topology, not to require a clean machine each time.

To remove the launchd services while leaving the Homebrew-managed binary installed:

```bash
ollama-logging-proxy-uninstall
```

## Local quality checks

Install [`just`](https://github.com/casey/just), `golangci-lint`, and `shellcheck`, then run:

```bash
just hooks   # one-time: enable commit-msg hook
just --list
just check
```

`just check` runs: `fmt`, `vet`, `lint`, `shellcheck`, `test`, and `race`.

## macOS scripts

Scripts are in [`scripts/`](scripts):

```bash
./scripts/install.sh
./scripts/smoke-test.sh
./scripts/uninstall.sh
```

`install.sh` is the source-checkout convenience path. It builds the proxy binary, then installs launch agents into `~/Library/LaunchAgents`, configures PRD env vars, and bootstraps both agents.

`install-launchd.sh` installs or updates the launchd setup for an already-installed binary, which is the right entrypoint for Homebrew installs. It is intended to be idempotent: unchanged healthy services should be left alone, while missing or unhealthy services should be reconciled back to the expected topology.

`smoke-test.sh` checks `GET /__ollama_logging_proxy/health`, sends `POST /api/generate`, and verifies today’s `body-YYYY-MM-DD.jsonl` contains a `/api/generate` entry.

`uninstall-launchd.sh` removes only the LaunchAgents and optionally the proxy logs.

`uninstall.sh` is the source-checkout convenience path. It removes the LaunchAgents and, by default, removes the locally installed binary. If the binary path points at Homebrew’s `bin`, it leaves the binary alone unless you explicitly set `REMOVE_BINARY=1`.

## CI

Pull requests run [`.github/workflows/ci.yml`](.github/workflows/ci.yml), which executes `just check`.
