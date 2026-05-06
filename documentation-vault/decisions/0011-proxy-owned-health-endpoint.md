---
status: "Accepted"
---

# 0011 — Health Endpoint at `/__ollama_logging_proxy/health` Is Proxy-Owned and Never Forwarded

## Context

PRD §11.8 and §14.4 define the proxy-owned health endpoint. PRD §14.4 states: "The endpoint must be impossible to confuse with an Ollama endpoint." The chosen path is `/__ollama_logging_proxy/health`.

The double-underscore prefix combined with the full service name makes it visually and structurally distinct from Ollama's `/api/` namespace. PRD §11.8 states: "This endpoint must not be forwarded to Ollama."

The `health` subcommand (PRD §10.1) calls this endpoint to check whether the proxy is alive. The same endpoint is referenced in the README smoke-test instructions and in `scripts/smoke-test.sh`. The proxy intercepts requests to this path before they reach the reverse-proxy forwarding logic — if the request reaches Ollama, it would return a 404, undermining the health check semantics.

## Decision

We will expose a proxy-owned health endpoint at `/__ollama_logging_proxy/health` that returns `{"ok":true,"service":"ollama-logging-proxy"}`. The proxy handles this path itself and never forwards it upstream.

## Consequences

- **Positive:** The health endpoint confirms the proxy is running, not just that Ollama is up. A 200 response from `/__ollama_logging_proxy/health` proves the proxy process is alive and listening.
- **Positive:** The `__` prefix and service-name path make collision with any current or future Ollama API path extremely unlikely.
- **Positive:** Operators and smoke tests can reliably distinguish "proxy is up" from "Ollama is up" by using different endpoints.
- **Trade-off:** minimal — see alternatives. The only cost is the path-interception logic that must run for every request before the forwarding decision.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Use Ollama's own `/api/version` as a health check | Would confirm Ollama is up, but not that the proxy is working. If the proxy crashes and Ollama is still running, `/api/version` would be unreachable anyway (proxy owns port 11434), so this doesn't usefully distinguish the failure mode. |
| Separate health port (e.g., a management port on 11436) | Adds a second listener, complicates plist configuration, and requires firewall/network rules for the additional port. PRD §14.4 did not require a separate port. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0001-reverse-proxy-not-fork]]
- [[0005-single-binary-subcommands]]
- [[0002-two-process-launchd-topology]]
