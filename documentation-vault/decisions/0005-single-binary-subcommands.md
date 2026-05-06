---
status: "Accepted"
---

# 0005 — Single Binary Exposing `serve / health / tail / purge` Subcommands

## Context

PRD §10 states: "The software should be distributed as a single binary that provides both the long-running proxy service and a small command-line interface for local operation and diagnostics." The LaunchAgent runs the binary via the `serve` subcommand; the operator uses `health`, `tail`, and `purge` for local diagnostics.

Commit `44527fa` (2026-05-04) introduced the initial `cmd/ollama-proxy/main.go` with subcommand dispatch, and commit `71e4cf8` added unit tests for all four CLI commands. The AGENTS.md project structure section documents the binary's four supported subcommands.

Having a single binary avoids distributing separate executables for service vs. CLI use. It also means environment variables from the LaunchAgent context are automatically available when the same binary is invoked for `health` or `tail` checks.

## Decision

We will ship a single binary (`ollama-logging-proxy`) that exposes four subcommands: `serve`, `health`, `tail`, and `purge`. No separate CLI tool or helper binary will be distributed.

## Consequences

- **Positive:** One binary to install, upgrade, and uninstall. Homebrew formula manages a single artifact.
- **Positive:** Subcommands share the same configuration layer (`internal/config`), so env-var defaults are consistent between the service and diagnostic commands.
- **Positive:** The `health` subcommand can be used directly by monitoring scripts or as a LaunchAgent `HealthCheck` program argument.
- **Trade-off:** All subcommands must be compiled together. Adding or changing a subcommand requires rebuilding and reinstalling the full binary, even for a minor CLI tweak.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Separate `ollama-logging-proxy-cli` binary | Doubles the distribution surface. Config sharing between service and CLI becomes an explicit coordination problem. Adds complexity to the Homebrew formula and install scripts. |
| Shell scripts for `health` / `tail` / `purge` | Reasonable for simple checks, but diverges from the compiled binary and loses type safety, test coverage, and consistent config loading. `tail` in particular benefits from the Go config package for log dir resolution. |
| Plugin/extension system | Over-engineering for a four-command CLI serving a single operator. PRD §10.2 defers future commands to a later version. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[execution-plan]]
- [[0004-go-stdlib-reverseproxy]]
- [[0009-filename-based-retention]]
- [[0011-proxy-owned-health-endpoint]]
