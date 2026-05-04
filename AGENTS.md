# Repository Guidelines

## Project Structure & Module Organization
`cmd/ollama-proxy/` is the CLI entrypoint. The binary supports `serve`, `health`, `tail [lines]`, and `purge`. Core packages live under `internal/`: `proxy/` handles forwarding and body capture, `config/` loads env-driven settings, and `logging/`, `redact/`, and `retention/` support log writing and cleanup. End-to-end tests live in `tests/`; package unit tests sit beside code as `*_test.go`. Operational assets live in `scripts/` and `launchd/`.

## Build, Test, and Development Commands
Use Go `1.22` as declared in [go.mod](go.mod). `just` is the main local entrypoint:

- `just --list`: show recipes.
- `just check`: run the full gate: formatting check, `go vet`, `golangci-lint`, tests, and race tests.
- `just test`: run `go test ./...`.
- `just race`: run `go test -race ./...`.
- `./scripts/install.sh`: build and install the proxy and launch agents.
- `./scripts/smoke-test.sh`: verify health, proxying, and JSONL logging.
- `./scripts/uninstall.sh`: unload agents and remove installed files.

## Coding Style & Naming Conventions
This repo is Go-first. Follow `gofmt` output exactly. Keep packages narrow and literal. Use `CamelCase` for exported names, `camelCase` for internal helpers, and concise identifiers such as `CaptureEvent` or `HealthPath`. Shell scripts should keep clear, action-oriented names and stay Bash-compatible. Run `golangci-lint` before opening a PR.

## Testing Guidelines
Use table-driven tests where they help; otherwise plain sequential tests are fine. Test names should be descriptive; add a behavior phrase when the subject alone is too vague. Prefer colocated unit tests for package behavior and `tests/` for cross-package flows. Changes to capture, config parsing, retention, or logging should ship with test updates. Run `just test` at minimum; run `just check` before review.

## Commit & Pull Request Guidelines
History currently uses short Conventional Commit-style subjects like `chore: initial code`. Keep using `<type>: <imperative summary>` such as `feat:`, `fix:`, or `chore:`. Pull requests are CI-gated by [`.github/workflows/ci.yml`](.github/workflows/ci.yml), which runs `just check`. If CI fails, start there. PRs should explain behavior changes, note config or launchd impact, and include local `just check` results. If install behavior changes, mention all affected scripts, including `uninstall.sh`.

## Security & Configuration Tips
Do not commit captured logs, generated `body-YYYY-MM-DD.jsonl` files, or machine-specific paths. Preserve the redaction contract in `internal/redact`: for valid JSON bodies, any field named `images`, case-insensitive and at any nesting depth, is replaced with `"[redacted]"`. Malformed or non-JSON bodies pass through unchanged. Do not widen or narrow that behavior casually.
