---
status: "Accepted"
---

# 0002 — Two-Process launchd Topology: Ollama Private 11435, Proxy Public 11434

## Context

PRD §5 defines the default process layout: LAN/localhost clients connect to the proxy on `0.0.0.0:11434`, which forwards to Ollama listening privately on `127.0.0.1:11435`. PRD §11.2 formalizes this as a hard requirement: port ownership must be unambiguous, and "no Ollama process should remain bound to `0.0.0.0:11434`."

Without this split, the proxy cannot bind port `11434` because Ollama would already own it. The split also provides a security benefit: Ollama's private port `11435` is bound to `127.0.0.1` and therefore unreachable from LAN clients (PRD §11.7).

PRD §11.3 and §11.4 require two separate LaunchAgent services: `dev.ollama.server` (runs Ollama with `OLLAMA_HOST=127.0.0.1:11435`) and `dev.ollama.logging-proxy` (runs the proxy on `0.0.0.0:11434`). Both were introduced in commit `44527fa` under the initial labels `com.joseph.ollama-private` and `com.joseph.ollama-proxy`, later renamed in commit `eb85955`.

## Decision

We will run two separate user-level launchd services: Ollama bound to `127.0.0.1:11435` and the proxy bound to `0.0.0.0:11434`. The proxy is the only process that serves the public Ollama-compatible port.

## Consequences

- **Positive:** Port ownership is unambiguous. `lsof -nP -iTCP:11434` shows only the proxy; `lsof -nP -iTCP:11435` shows only Ollama.
- **Positive:** Ollama's private port is not reachable from LAN clients by default, because `127.0.0.1` does not accept external connections.
- **Positive:** Each service can be restarted independently without disrupting the other.
- **Trade-off:** A setup operator must reconfigure an existing Ollama LaunchAgent to use `OLLAMA_HOST=127.0.0.1:11435`. If Ollama is still bound to `11434` at install time, the proxy cannot start. The install scripts handle this, but it is a real migration step.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Single process (proxy embeds Ollama) | Ollama is a complex application with its own lifecycle; embedding it would reintroduce the fork/patch problem. Rejected for the same reasons as ADR 0001. |
| Proxy on a different port (e.g., 11436) | Clients would need reconfiguration. PRD §3 explicitly requires existing clients to keep using `11434` without code changes. |
| Proxy and Ollama share port via SO_REUSEPORT | Two competing processes on the same socket is not a standard pattern for this use case, and does not provide body-logging visibility. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0001-reverse-proxy-not-fork]]
- [[0003-user-launchagent-scope]]
- [[0015-launchagent-labels-com-joseph]]
- [[0016-launchagent-labels-dev-ollama]]
