---
status: "Accepted"
---

# 0016 — Rename LaunchAgent Labels to `dev.ollama.logging-proxy` and `dev.ollama.server`

## Context

ADR 0015 documents the original labels (`com.joseph.ollama-proxy` and `com.joseph.ollama-server`) introduced in commit `44527fa`. Those labels used a personal namespace and an inconsistent service name for the Ollama process (`ollama-private` vs. what PRD §11.2 calls `dev.ollama.server`).

Commit `eb85955` (2026-05-05, "fix(launchd): harden installer and rename labels") renamed the plist files:
- `com.joseph.ollama-proxy.plist` → `dev.ollama.logging-proxy.plist`
- `com.joseph.ollama-private.plist` → `dev.ollama.server.plist`

PRD §11.2 (as updated) names the two services `dev.ollama.server` and `dev.ollama.logging-proxy`. The `dev.ollama.*` namespace is project-scoped, matches the public Ollama developer domain conventions, and clearly communicates that these are developer/local services related to the Ollama project. The same commit also hardened the installer scripts significantly.

## Decision

We will use `dev.ollama.logging-proxy` and `dev.ollama.server` as the canonical LaunchAgent labels. These supersede the initial `com.joseph.*` labels from ADR 0015.

## Consequences

- **Positive:** The `dev.ollama.*` namespace is project-scoped and communicates the relationship to Ollama without implying Ollama project ownership. Multiple contributors can deploy the same service without label conflicts.
- **Positive:** Service names (`logging-proxy`, `server`) describe function, not port binding, making them clearer than the old `ollama-private` label.
- **Positive:** PRD §11.2 now uses the new labels in all topology diagrams, creating consistency between the PRD, plists, and install scripts.
- **Trade-off:** Any existing installation using the old `com.joseph.*` labels requires an unload + reload cycle. The hardened installer in commit `eb85955` handles this migration, but it is a breaking change for anyone who installed from the initial code commit.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Keep `com.joseph.*` labels | Personal namespace is not project-scoped. PRD §11.2 had already adopted the `dev.ollama.*` naming, creating a documentation/code mismatch. |
| `com.ollama.logging-proxy` (official-looking namespace) | Could be confused with an official Ollama project service. `dev.ollama.*` clearly connotes a developer/local tool in the Ollama ecosystem. |

## Supersedes

[[0015-launchagent-labels-com-joseph]]

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0002-two-process-launchd-topology]]
- [[0015-launchagent-labels-com-joseph]]
