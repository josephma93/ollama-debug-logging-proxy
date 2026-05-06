---
status: "Accepted"
---

# 0013 — Conventional Commit Subjects Enforced via Local commit-msg Hook and CI

## Context

AGENTS.md "Commit & Pull Request Guidelines" section specifies the Conventional Commit format: `<type>(<scope>)?!?: <imperative summary>` with a defined set of allowed types (`feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `ci`, `build`, `perf`, `revert`, `style`, `deps`). It notes that the format rule lives in `scripts/check-commit-subject.sh` and is enforced by both a local hook and CI.

Commit `13d1263` (2026-05-05) introduced the full enforcement infrastructure: `scripts/check-commit-subject.sh` as the single source of truth for the validation regex, `.githooks/commit-msg` as the local client-side hook, and an updated `.github/workflows/ci.yml` that re-validates commit subjects on PRs (rejecting `fixup!`/`squash!` markers to force autosquash before push).

The commit message for `13d1263` describes the design explicitly: "Belt-and-suspenders enforcement of AGENTS.md rules at both layers." The `just hooks` recipe wires up the local hook via `git config core.hooksPath .githooks/`.

## Decision

We will enforce Conventional Commit subject format at two checkpoints: a local `commit-msg` hook (enabled via `just hooks`) and a CI job that validates all PR commits against the same script. The validator is `scripts/check-commit-subject.sh`, which is the single source of truth for the format rule.

## Consequences

- **Positive:** Commit history is machine-parseable. Future changelog generation, semantic versioning automation, or release note tooling can rely on the type prefix.
- **Positive:** The dual-layer enforcement catches violations either locally (fast feedback) or at CI time (for contributors who skip `just hooks`).
- **Positive:** `fixup!`/`squash!`/`amend!` markers are tolerated locally but rejected in CI, encouraging clean history before merge.
- **Trade-off:** Contributors must run `just hooks` once after cloning or face CI failure on their first PR. This is documented in AGENTS.md and README but is an extra setup step.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Commit lint via npm / commitlint | Adds a Node.js dependency to a Go project. The project already has `scripts/` as a shell convention; a shell-based validator fits the existing tooling model. |
| CI-only enforcement (no local hook) | Slower feedback loop: the developer only learns about format violations after pushing. The local hook gives instant feedback without CI round-trip. |
| No enforcement (convention by documentation only) | Conventions without gates drift over time. The existing commit history shows mixed styles before `13d1263`; the gate was introduced to prevent further drift. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[execution-plan]]
- [[0004-go-stdlib-reverseproxy]]
