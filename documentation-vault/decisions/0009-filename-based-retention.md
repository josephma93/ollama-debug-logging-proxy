---
status: "Accepted"
---

# 0009 — Retention Parses the Date from `body-YYYY-MM-DD.jsonl`, Ignoring mtime

## Context

PRD §7.9 requires that the proxy automatically delete body log files older than the configured retention period (default 10 days). PRD §14.3 adds a product decision: "Log retention must be based on the date embedded in the filename, not file modification time." It also specifies that files not matching the expected pattern must be ignored by retention cleanup.

The rationale is operational safety: file modification time (`mtime`) can change for reasons unrelated to the file's logical date (backup restoration, filesystem migration, `touch` commands, time zone changes). Parsing the date from `body-2026-05-04.jsonl` gives a stable, predictable retention boundary.

Commit `44527fa` introduced `internal/retention/retention.go` and `internal/retention/retention_test.go` implementing filename-based retention. PRD §7.9 states cleanup runs at startup and no more than once per hour during writes.

## Decision

We will determine a log file's age by parsing the `YYYY-MM-DD` portion of its filename, not by reading the file's `mtime`. Files whose filename date is older than `OLLAMA_PROXY_RETENTION_DAYS` are deleted; files not matching the pattern `body-YYYY-MM-DD.jsonl` are left untouched.

## Consequences

- **Positive:** Retention is deterministic and reproducible. An operator can predict which files will be deleted by looking at filenames, not filesystem metadata.
- **Positive:** Safe across filesystem migrations, backup restorations, and `touch` operations — none of these change the filename.
- **Positive:** Files with non-matching names (e.g., `stdout.log`, `stderr.log`, operator-created notes) are ignored, preventing accidental deletion.
- **Trade-off:** If a log file is renamed to a different date (e.g., manually moved), retention uses the new name's date. This is an edge case unlikely in normal operation but worth noting.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| mtime-based retention | mtime is unreliable after filesystem operations (restores, syncs, `touch`). PRD §14.3 explicitly requires filename-based retention to avoid this. |
| Size-based rotation (logrotate-style) | Rotates on file size, not age. Produces files with unpredictable dates, breaking the daily-file model. Does not implement the time-based retention the PRD requires. |
| SQLite database with TTL field | Would allow row-level expiry but abandons the simple JSONL file format and adds a database dependency. PRD §7.8 specifies daily JSONL files as the canonical format. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0010-jsonl-daily-files]]
- [[0005-single-binary-subcommands]]
- [[0006-env-var-only-config]]
