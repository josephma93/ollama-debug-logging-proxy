---
status: "Accepted"
---

# 0010 — Body Logs Are Daily JSONL Files with Serialized (Mutex) Writes

## Context

PRD §7.3.1 defines the canonical JSONL record shape: one JSON object per line, with fields for request metadata, request body, response body, truncation flags, and redaction flags. PRD §7.8 specifies daily file naming (`body-YYYY-MM-DD.jsonl`) and a default log directory of `~/Library/Logs/ollama-proxy`.

PRD §12 (Proposed Implementation) states: "Daily JSONL logging should be handled internally by the proxy. A mutex should protect writes so concurrent requests do not interleave log entries."

Commit `44527fa` introduced `internal/logging/logging.go` implementing a mutex-protected, append-only daily file writer. The execution plan §2.3 (Task 2.3) specifies JSONL as the format and daily rotation as the file strategy.

The format must support both `tail`-style inspection (human-readable via CLI) and potential future export or analysis tooling (PRD §10.2 mentions an `export` command). JSONL is line-oriented and grep-friendly.

## Decision

We will write body logs as append-only daily JSONL files (`body-YYYY-MM-DD.jsonl`). Concurrent writes are serialized via a mutex to prevent interleaved log entries.

## Consequences

- **Positive:** Each line is a self-contained JSON object. Files can be read with standard tools (`cat`, `jq`, `grep`) without special parsers.
- **Positive:** Daily rotation naturally partitions logs for retention cleanup and makes it easy to find entries by date.
- **Positive:** A mutex is sufficient for the single-process, single-machine deployment target. There are no distributed write concerns.
- **Trade-off:** The mutex serializes all log writes. Under very high concurrent request rates, this could become a bottleneck. For the intended single-user local debug use case, this is not a practical concern.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Single growing log file (no daily rotation) | Makes filename-based retention (ADR 0009) impossible. A single file grows indefinitely and is harder to inspect by date. |
| SQLite database | Enables rich queries and row-level TTL but adds a database dependency and loses the simplicity of grep-able text files. PRD §7.8 specifies JSONL as the format. |
| Parquet or columnar format | Efficient for analytics but requires a reader library; not inspectable with `cat` or `jq`. Excessive for a local debug tool. |
| OpenTelemetry export | Requires an OTLP collector/backend. PRD §4 explicitly excludes OpenTelemetry integration from the first version. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0007-tapped-endpoint-allowlist]]
- [[0008-log-only-image-redaction]]
- [[0009-filename-based-retention]]
- [[0012-omit-headers-from-body-logs]]
