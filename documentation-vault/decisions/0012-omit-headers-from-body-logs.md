---
status: "Accepted"
---

# 0012 — Request Headers Are Intentionally Omitted from Body Log Records

## Context

PRD §14.5 states: "Request headers should be intentionally omitted from body logs in this version. The log should remain focused on request metadata, request body, response body, status, timing, truncation, and redaction indicators."

PRD §7.3.1 defines the canonical JSONL schema. Examining the schema, there is no `headers` or `request_headers` field. The schema includes `user_agent` (a single extracted header value) and `client_ip` (derived from the connection), but the full header map is absent by design.

The rationale is twofold: headers can contain sensitive values (Authorization tokens, API keys, cookies, custom authentication headers), and they add log volume without contributing to the primary debugging goal of understanding what prompts and responses were exchanged. Commit `44527fa` implemented `internal/logging/logging.go` with the §7.3.1 schema; no header field was included.

PRD §13 (Future Enhancements) notes that "optional header logging" may be added later.

## Decision

We will not include request headers in JSONL body log records in the first version. The `user_agent` string is the only header-derived field logged, extracted explicitly rather than captured as part of the header map.

## Consequences

- **Positive:** Log files do not capture Authorization headers, cookies, or other sensitive header values that may be present in client requests to Ollama.
- **Positive:** Log records are more compact; header maps can add several kilobytes per request for clients that send many headers.
- **Trade-off:** If an operator wants to debug authentication or routing behavior driven by custom headers, the body logs are insufficient and the operator must consult Ollama's own logs or use a separate network inspection tool.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Log all request headers | Risk of capturing `Authorization`, `Cookie`, and other sensitive values. PRD §14.5 explicitly defers this to a future version. |
| Log a configurable subset of headers | Adds configuration complexity (which headers to allow/deny) and requires a new env var or config mechanism. Deferred per PRD §14.5. |
| Log response headers | Not mentioned in PRD §7.3.1 schema and not required for the observability goal of the first version. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0007-tapped-endpoint-allowlist]]
- [[0008-log-only-image-redaction]]
- [[0010-jsonl-daily-files]]
