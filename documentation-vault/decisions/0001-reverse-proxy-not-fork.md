---
status: "Accepted"
---

# 0001 — Reverse Proxy in Front of Ollama, Don't Fork or Patch

## Context

PRD §1 describes the project as "a lightweight HTTP reverse proxy that sits between local or LAN clients and a private Ollama server." The core problem (PRD §2) is that Ollama's built-in logs do not expose request or response bodies, creating an observability gap for debugging local applications.

Three approaches were available to gain body-level visibility: modify Ollama itself (fork or patch), instrument every calling client, or place a transparent intermediary in front of an unmodified Ollama process. PRD §3 states the proxy "should not require Ollama-specific client changes" and "Ollama itself should move behind the proxy." The initial PRD was committed as `7d7d9bc` (2026-05-04), and initial code in `44527fa` implemented this as a Go reverse proxy with no Ollama modifications.

PRD §12 (Acceptance Criteria) item 1 requires that "existing clients can call `http://<host>:11434` without changing their configuration," which a fork or patch approach would not satisfy without an equivalent proxy layer anyway.

## Decision

We will implement observability as a transparent HTTP reverse proxy that fronts an unmodified Ollama process, not by forking, patching, or instrumenting Ollama or its clients.

## Consequences

- **Positive:** Ollama upgrades are independent — the proxy works with any Ollama version that preserves the REST API surface. No Ollama fork to maintain or rebase.
- **Positive:** Existing clients need zero reconfiguration; they continue pointing at port `11434`.
- **Positive:** The proxy is a thin infrastructure layer: it adds latency only from body capture overhead, and non-tapped endpoints incur no body-logging cost.
- **Trade-off:** The proxy cannot observe traffic that bypasses it (e.g., a client connecting directly to port `11435`). Port discipline must be maintained externally (PRD §11.5).

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Fork or patch Ollama | Requires maintaining a divergent Ollama codebase and rebasing every upstream release. Directly conflicts with PRD §3's goal of leaving Ollama unmodified. |
| Instrument each calling client | Every client (scripts, apps, LAN devices) must be modified individually. Breaks the PRD requirement that existing clients work without changes. |
| Sidecar logging plugin / shared memory tap | Ollama has no plugin API. A shared-memory or eBPF approach would be OS-specific, fragile across Ollama versions, and far out of scope for a local debug tool. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0002-two-process-launchd-topology]]
- [[0003-user-launchagent-scope]]
- [[0011-proxy-owned-health-endpoint]]
