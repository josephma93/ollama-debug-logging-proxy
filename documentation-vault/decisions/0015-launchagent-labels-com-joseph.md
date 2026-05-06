---
status: "Superseded"
---

# 0015 — Initial LaunchAgent Labels `com.joseph.ollama-proxy` and `com.joseph.ollama-server`

## Context

Commit `44527fa` (2026-05-04) introduced the initial launchd plist files as `launchd/com.joseph.ollama-proxy.plist` and `launchd/com.joseph.ollama-private.plist`. These labels followed the macOS reverse-DNS convention but used a personal namespace (`com.joseph`) and inconsistent naming (`ollama-private` for the server vs. `ollama-proxy` for the proxy).

At the time of the initial code commit, the two-process topology was already decided (see ADR 0002), but the naming convention for the launchd labels had not been finalized. The `com.joseph` prefix was a quick personal namespace choice, and `ollama-private` was a functional description of the port binding rather than a service identity.

## Decision

We used the LaunchAgent labels `com.joseph.ollama-proxy` and `com.joseph.ollama-server` as the initial identifiers for the proxy and Ollama services.

## Consequences

- **Positive:** Labels followed the macOS reverse-DNS convention structure.
- **Trade-off:** The `com.joseph` namespace is personal, not project-scoped. It would conflict with any other user's deployment and does not communicate the project identity. `ollama-private` as a label name was confusing — it described the port, not the service.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| `com.josephma93.ollama-proxy` (GitHub username namespace) | More unique but still a personal namespace. Not chosen at initial commit — superseded by a project-scoped `dev.ollama.*` namespace instead. |
| `io.ollama.*` (official Ollama namespace) | Would conflict with any future official Ollama launchd labels and implies Ollama project ownership. Not appropriate for a third-party tool. |

## Supersedes

## Superseded by

[[0016-launchagent-labels-dev-ollama]]

## Related

- [[vault-conventions]]
- [[prd]]
- [[0002-two-process-launchd-topology]]
- [[0016-launchagent-labels-dev-ollama]]
