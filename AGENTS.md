# Repository Guidelines

## Project Structure & Module Organization
`cmd/ollama-proxy/` is the CLI entrypoint. The binary supports `serve`, `health`, `tail [lines]`, and `purge`. Core packages live under `internal/`: `proxy/` handles forwarding and body capture, `config/` loads env-driven settings, and `logging/`, `redact/`, and `retention/` support log writing and cleanup. End-to-end tests live in `tests/`; package unit tests sit beside code as `*_test.go`. Operational assets live in `scripts/` and `launchd/`.

## Build, Test, and Development Commands
Use Go `1.22` as declared in [go.mod](go.mod). `just` is the main local entrypoint:

- `just --list`: show recipes.
- `just hooks`: one-time setup — points `core.hooksPath` at `.githooks/` so the commit-msg gate runs locally. Run once after cloning.
- `just check`: run the full gate: formatting check, `go vet`, `golangci-lint`, `shellcheck`, tests, and race tests.
- `just test`: run `go test ./...`.
- `just race`: run `go test -race ./...`.
- `./scripts/install.sh`: build and install the proxy and launch agents.
- `./scripts/smoke-test.sh`: verify health, proxying, and JSONL logging.
- `./scripts/uninstall.sh`: unload agents and remove installed files.

Required tools: Go `1.22`, [`just`](https://github.com/casey/just), `golangci-lint`, `shellcheck`.

## Coding Style & Naming Conventions
This repo is Go-first. Follow `gofmt` output exactly. Keep packages narrow and literal. Use `CamelCase` for exported names, `camelCase` for internal helpers, and concise identifiers such as `CaptureEvent` or `HealthPath`. Shell scripts should keep clear, action-oriented names and stay Bash-compatible. Run `golangci-lint` before opening a PR.

## Testing Guidelines
Use table-driven tests where they help; otherwise plain sequential tests are fine. Test names should be descriptive; add a behavior phrase when the subject alone is too vague. Prefer colocated unit tests for package behavior and `tests/` for cross-package flows. Changes to capture, config parsing, retention, or logging should ship with test updates. Run `just test` at minimum; run `just check` before review.

## Commit & Pull Request Guidelines
History uses short Conventional Commit-style subjects like `chore: initial code`. Keep using `<type>(<scope>)?!?: <imperative summary>`, where `<type>` is one of `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `ci`, `build`, `perf`, `revert`, `style`, `deps`. The format rule lives in [`scripts/check-commit-subject.sh`](scripts/check-commit-subject.sh); the local `commit-msg` hook (`.githooks/commit-msg`, enabled by `just hooks`) and the CI workflow both call it, so a missed `just hooks` only delays the failure to PR time. The local hook tolerates `fixup!`/`squash!`/`amend!` markers; CI does not, so autosquash before pushing. Pull requests are CI-gated by [`.github/workflows/ci.yml`](.github/workflows/ci.yml), which validates commit subjects and then runs `just check`. If CI fails, start there. PRs should explain behavior changes, note config or launchd impact, and include local `just check` results. If install behavior changes, mention all affected scripts, including `uninstall.sh`.

## Security & Configuration Tips
Do not commit captured logs, generated `body-YYYY-MM-DD.jsonl` files, or machine-specific paths. Preserve the redaction contract in `internal/redact`: for valid JSON bodies, any field named `images`, case-insensitive and at any nesting depth, is replaced with `"[redacted]"`. Malformed or non-JSON bodies pass through unchanged. Do not widen or narrow that behavior casually.
