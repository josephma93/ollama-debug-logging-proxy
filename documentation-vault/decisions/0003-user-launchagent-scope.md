---
status: "Accepted"
---

# 0003 — macOS User LaunchAgent (Not LaunchDaemon, Not brew services)

## Context

PRD §5 and §11.1 specify that the initial supported environment is macOS, with both the proxy and Ollama running as user-level LaunchAgent services under the same user account that owns the Ollama model directory and log directory. PRD §11.1 states "both services should run under the same macOS user account."

macOS provides several service management mechanisms: user-level `LaunchAgent` (runs as the logged-in user, under `~/Library/LaunchAgents`), system-level `LaunchDaemon` (runs as root or a system user, under `/Library/LaunchDaemons`), and `brew services` (Homebrew's abstraction over launchd). The README "Installer Contract" section explicitly states "it does not use `brew services`" and "it does not install anything into `/Library/LaunchDaemons`."

The log directory (`~/Library/Logs/ollama-proxy`) and Ollama model directory (`~/.ollama/models`) are both under the user home, making user-level service scope the natural fit.

## Decision

We will run both services as macOS user LaunchAgents (under `~/Library/LaunchAgents`), managed by the user's login session, not as system daemons and not via `brew services`.

## Consequences

- **Positive:** Services run with the user's identity, giving them natural access to `~/Library/Logs/`, `~/.ollama/`, and `~/bin/` without elevated permissions or sudo.
- **Positive:** Services start automatically at user login and are scoped to the user session, matching Ollama's own macOS behavior.
- **Positive:** No root/system privilege required for install, uninstall, or service management.
- **Trade-off:** Services do not run when no user is logged in (e.g., headless macOS server). If always-on behavior without a user session is needed, a LaunchDaemon approach would be required, but that is explicitly out of scope for the current target environment.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| `brew services` | `brew services` manages launchd plists on behalf of Homebrew but adds indirection and limits fine-grained control over environment variables. The install scripts need to set specific env vars (e.g., `OLLAMA_HOST`) and the README explicitly excludes this approach. |
| System LaunchDaemon (`/Library/LaunchDaemons`) | Runs as root or a dedicated system user, requires elevated install privileges, and separates the service identity from the user account that owns the log and model directories. Rejected per PRD §11.1 and the README installer contract. |
| `launchctl submit` (deprecated API) | Deprecated in modern macOS. Not an option for new services. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0001-reverse-proxy-not-fork]]
- [[0002-two-process-launchd-topology]]
- [[0014-homebrew-tap-stable-canary-channels]]
