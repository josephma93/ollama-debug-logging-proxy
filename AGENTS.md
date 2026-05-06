# Repository Guidelines

## Project Structure & Module Organization
`cmd/ollama-proxy/` is the CLI entrypoint. The binary supports `serve`, `health`, `tail [lines]`, and `purge`. Core packages live under `internal/`: `proxy/` handles forwarding and body capture, `config/` loads env-driven settings, and `logging/`, `redact/`, and `retention/` support log writing and cleanup. End-to-end tests live in `tests/`; package unit tests sit beside code as `*_test.go`. Operational assets live in `scripts/` and `launchd/`. Long-form project documentation lives in `documentation-vault/` — see [Documentation Vault](#documentation-vault) below before touching it.

## Documentation Vault
Long-form project documentation lives in [`documentation-vault/`](documentation-vault/home.md). Before adding, moving, splitting, or deleting any vault note, read [`documentation-vault/vault-conventions.md`](documentation-vault/vault-conventions.md) — **mandatory**. The vault has 22 numbered hard rules (R1–R22) and they block merge.

Critical things to know before you touch the vault:

- **Bucket placement.** Every note belongs in exactly one of `explanation/` (the why), `reference/` (the what), `how-to/` (the steps), `decisions/` (ADRs, append-only), or `releases/` (per-tag). The vault root holds only `home.md`, `vault-conventions.md`, and `glossary.md`. Adding a new top-level folder requires an ADR (R1, R21).
- **Wikilinks are short form.** `[[name]]`, not `[[bucket/name]]`. Filenames are unique by R4. The only path-qualified exceptions are `[[decisions/index]]` and `[[releases/_template]]`.
- **Every note ends with `## Related`.** That is what makes the vault a network rather than a folder of orphans.
- **ADRs are append-only.** To change an accepted decision, write a new numbered ADR that supersedes the old one — do not edit the old one in place. See the 0015 ↔ 0016 worked example in [`documentation-vault/decisions/`](documentation-vault/decisions/index.md).
- **Adding `decisions/NNNN-*.md`?** Update [`decisions/index.md`](documentation-vault/decisions/index.md) in the same commit (R11).
- **Pushing a release tag?** Add `releases/<tag>.md` first (R12).
- **No prose duplication between code-rooted docs and the vault.** This file and [`README.md`](README.md) describe **current code state**; the vault explains **thinking**. Link, do not duplicate (R17).
- **Hygiene.** No `TODO` / `FIXME` / `XXX` markers in committed vault notes (R20). No hardcoded user paths or LAN IPs (R19) — use `$HOME`, `<user>`, `<lan-ip>`.
- **Changes to `vault-conventions.md`** require a new ADR proposing the change, in the same commit (R21). The hard-rules block was itself adopted via [ADR 0017](documentation-vault/decisions/0017-hard-rules-for-vault-hygiene.md).

[`vault-conventions.md`](documentation-vault/vault-conventions.md) is authoritative. Its `## Hard rules` section answers "where does this go?" and "is this allowed?" definitively; the upper sections explain why the rules look the way they do.

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
