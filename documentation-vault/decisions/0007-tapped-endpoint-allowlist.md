---
status: "Accepted"
---

# 0007 — Body Logging Only for Tapped Inference Endpoints

## Context

PRD §7.2 is titled "Normative body logging scope" and is explicitly the authoritative definition of which requests receive body-level logging. It states the proxy must proxy all Ollama API endpoints transparently, but must only "capture, redact, truncate, and store request and response bodies" for four specific inference endpoints:

```
/api/generate
/api/chat
/api/embeddings
/api/embed
```

PRD §14.2 restates this as a product decision: "The proxy must not body-log every Ollama endpoint." The rationale is that non-inference endpoints such as `/api/tags` contain no sensitive prompts or responses, and body-logging them would add noise and unnecessary storage without contributing to the observability goal.

PRD §7.2 also specifies that path matching is exact (normalized path) and query strings do not affect tapped status: `/api/chat?x=1` is tapped because its path is `/api/chat`. Commit `44527fa` introduced `internal/proxy/classifier.go` which implements this allowlist logic.

## Decision

We will body-log only the four inference endpoints listed above. All other Ollama endpoints are proxied transparently without body capture or JSONL log entries.

## Consequences

- **Positive:** Non-inference endpoints (`/api/tags`, `/api/version`, `/api/show`, etc.) produce zero logging overhead — no buffer allocation, no file I/O.
- **Positive:** Log files remain focused on inference traffic, making them easier to read and analyze.
- **Positive:** The allowlist is explicit and tested. Adding a new endpoint to the list is a deliberate, code-visible change.
- **Trade-off:** Endpoints added to the Ollama API in future versions are not tapped by default. An operator who wants to observe a new inference endpoint must update the allowlist. PRD §7.2 acknowledges this: "current and future endpoints" are proxied, but not body-logged unless explicitly tapped.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Log all proxied requests | Produces large volumes of metadata-only log noise for non-inference calls. `/api/tags` is called frequently by many clients for model discovery and contributes nothing to prompt debugging. PRD §14.2 explicitly rejects this. |
| Denylist (log everything except excluded paths) | Inverts the control: new Ollama endpoints are logged by default until someone adds them to a denylist. The allowlist approach is safer for privacy and storage. |
| Pattern-based matching (e.g., `/api/*`) | `/api/tags`, `/api/version`, `/api/show`, `/api/copy`, `/api/delete`, `/api/pull`, `/api/push` are all under `/api/` but are non-inference endpoints. A wildcard would capture all of them. Explicit enumeration is more precise. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0008-log-only-image-redaction]]
- [[0010-jsonl-daily-files]]
- [[0012-omit-headers-from-body-logs]]
