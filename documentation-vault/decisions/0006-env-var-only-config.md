---
status: "Accepted"
---

# 0006 — Configuration via Environment Variables Only, No Config File

## Context

PRD §8.4 (Operability) states: "The proxy must be configurable through environment variables so LaunchAgent deployment remains simple. Configuration should not require a separate config file in the first version."

PRD §9 defines the five supported environment variables (`OLLAMA_PROXY_LISTEN`, `OLLAMA_PROXY_TARGET`, `OLLAMA_PROXY_LOG_DIR`, `OLLAMA_PROXY_RETENTION_DAYS`, `OLLAMA_PROXY_MAX_BODY_BYTES`) with explicit defaults. The launchd plist templates set all five vars directly in the `EnvironmentVariables` dictionary, making the plist the canonical configuration document for the deployed service.

Commit `44527fa` (2026-05-04) implemented `internal/config` as a pure env-var reader with no file-based fallback. The `internal/config` package was present from the first code commit and has remained a thin env-var loader.

## Decision

We will configure the proxy exclusively through environment variables. There is no config file, no TOML/YAML/JSON configuration, and no flag-file mechanism in the first version.

## Consequences

- **Positive:** LaunchAgent plists double as the authoritative configuration document. No separate config file to track, version, or migrate.
- **Positive:** Environment variable configuration is trivially inspectable via `launchctl print gui/<uid>/dev.ollama.logging-proxy`.
- **Positive:** All subcommands inherit the same environment, so `health`, `tail`, and `purge` automatically use the same `OLLAMA_PROXY_LOG_DIR` as `serve` when invoked in the same shell.
- **Trade-off:** Changing a config value requires editing the plist and reloading the LaunchAgent. There is no hot-reload mechanism. For a single-user local debug tool, this is acceptable, but it would be a friction point for multi-service or multi-user deployments.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| TOML or YAML config file | Adds a config file path convention, parsing dependency, and file-finding logic. PRD §8.4 explicitly excludes this for the first version. |
| CLI flags (cobra/flag package) | Flags are redundant when env vars are already the mechanism for LaunchAgent deployment. Two config surfaces for five variables adds complexity without benefit. |
| Mixed (flags override env vars) | Reasonable pattern for production CLIs, but the proxy is a single-user service. Deferred to a potential future version per PRD §8.4. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0002-two-process-launchd-topology]]
- [[0003-user-launchagent-scope]]
- [[0009-filename-based-retention]]
