# Ollama Logging Proxy

Reverse proxy in front of Ollama that preserves normal API behavior and writes tapped request/response bodies to daily JSONL logs.

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

This repo now includes a Homebrew formula at [`Formula/ollama-logging-proxy.rb`](Formula/ollama-logging-proxy.rb).

Current status:

- Before the first tagged GitHub release, install from source with Homebrew:

```bash
brew install --HEAD ./Formula/ollama-logging-proxy.rb
```

- After release automation has published a tagged release and updated the formula, install from the tap:

```bash
brew tap josephma93/ollama-debug-logging-proxy https://github.com/josephma93/ollama-debug-logging-proxy
brew install josephma93/ollama-debug-logging-proxy/ollama-logging-proxy
```

Homebrew only installs the binary and helper assets. It does not automatically rewrite your launchd topology.

To wire the proxy into launchd after `brew install`:

```bash
ollama-logging-proxy-install
```

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

`install.sh` is the source-checkout convenience path. It builds the proxy binary, then installs launch agents into `~/Library/LaunchAgents`, configures PRD env vars, and bootstraps both agents.

`install-launchd.sh` installs or updates the launchd setup for an already-installed binary, which is the right entrypoint for Homebrew installs.

`smoke-test.sh` checks `GET /__ollama_logging_proxy/health`, sends `POST /api/generate`, and verifies today’s `body-YYYY-MM-DD.jsonl` contains a `/api/generate` entry.

`uninstall-launchd.sh` removes only the LaunchAgents and optionally the proxy logs.

`uninstall.sh` is the source-checkout convenience path. It removes the LaunchAgents and, by default, removes the locally installed binary. If the binary path points at Homebrew’s `bin`, it leaves the binary alone unless you explicitly set `REMOVE_BINARY=1`.

## Release Automation

Homebrew release automation follows this sequence:

Tag semantics:

- Stable release tag: `v0.1.0`
- Prerelease tag: `v0.1.1-canary.1`, `v0.1.1-rc.1`, `v0.1.1-beta.1`, `v0.1.1-alpha.1`
- Rule: if the version contains a hyphen suffix after the numeric core, it is treated as a prerelease.

Homebrew release automation follows this sequence:

1. Push a version tag such as `v0.1.0` or `v0.1.1-canary.1`.
2. [`.github/workflows/release.yml`](.github/workflows/release.yml) builds macOS release tarballs for `arm64` and `x86_64`, including:
   - `ollama-logging-proxy`
   - `scripts/install-launchd.sh`
   - `scripts/uninstall-launchd.sh`
   - `launchd/*.plist`
3. The workflow uploads those tarballs and matching `.sha256` files to the GitHub Release.
4. If the tag is a stable tag, the same release workflow regenerates [`Formula/ollama-logging-proxy.rb`](Formula/ollama-logging-proxy.rb) from [`.github/formula-template.rb`](.github/formula-template.rb) using the release checksums and pushes the formula update back to `main`.
5. If the tag is a prerelease tag, the workflow still publishes release artifacts, but it does not rewrite the stable Homebrew formula.

## CI

Pull requests run [`.github/workflows/ci.yml`](.github/workflows/ci.yml), which executes `just check`.
