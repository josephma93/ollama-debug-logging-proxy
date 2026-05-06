---
status: "Accepted"
---

# 0008 — Image Redaction Applies Only to the Log Copy, Never to Upstream Traffic

## Context

PRD §3 states: "The proxy must redact image payloads entirely from logs. Any JSON field named `images`, at any nesting depth, should be replaced before writing the request or response body to disk." The motivation is to prevent base64-encoded image payloads from making logs excessively large and unreadable.

PRD §7.7 defines the redaction behavior in detail: recursive depth, case-insensitive field name match, replacement with the string `"[redacted]"`. Crucially, PRD §7.4 states: "The proxy must not alter the actual request body sent to Ollama, except for normal proxy forwarding behavior. Redaction applies only to the log copy, not to traffic forwarded upstream."

This constraint is non-obvious: the proxy must read the request body, produce a redacted copy for logging, and restore the original (unredacted) body for forwarding. Commit `44527fa` introduced `internal/redact/redact.go` and `internal/proxy/capture.go`, which implement this split: captured bytes are redacted before writing to the log, while the original reader is passed unmodified to the upstream.

AGENTS.md also notes: "Preserve the redaction contract in `internal/redact`... Do not widen or narrow that behavior casually."

## Decision

We will apply image redaction exclusively to the in-memory log copy. The bytes forwarded to Ollama and returned to the client are never modified by the redaction logic.

## Consequences

- **Positive:** Ollama receives the exact request the client sent. Multimodal models using the `images` field continue to work correctly through the proxy.
- **Positive:** Log files stay readable and compact — a 1 MB base64 image becomes `"[redacted]"` in the log while the real request is processed normally.
- **Trade-off:** The proxy must buffer a copy of the request body in memory before forwarding it. For requests with large image payloads, this means holding up to `OLLAMA_PROXY_MAX_BODY_BYTES` of the body in memory simultaneously with the in-flight request. The body size cap (PRD §7.6) limits this exposure.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Redact the body before forwarding to Ollama | Breaks multimodal inference. Ollama would receive a request with `"images": "[redacted]"` instead of actual image data, causing model errors. Explicitly excluded by PRD §7.4. |
| Skip logging entirely for requests with image fields | Loses all observability for multimodal requests. The prompt, model name, and response would not be logged. Violates the core observability goal. |
| Size-based truncation instead of field redaction | Truncation at a byte limit could cut off mid-JSON, producing unparseable log entries. Field-specific redaction preserves JSON structure while eliminating the large payload. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0007-tapped-endpoint-allowlist]]
- [[0010-jsonl-daily-files]]
- [[0012-omit-headers-from-body-logs]]
