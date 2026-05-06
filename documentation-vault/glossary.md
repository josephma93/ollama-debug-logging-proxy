# Glossary

Shared term definitions for the Ollama Logging Proxy project. When a term appears in multiple notes, link to this file rather than re-defining it.

---

**tapped request** — A proxied HTTP request whose normalized URL path exactly matches one of the four inference endpoints (`/api/generate`, `/api/chat`, `/api/embeddings`, `/api/embed`). Only tapped requests have their bodies captured, redacted, and stored in body logs. Query strings do not affect tapping. Defined normatively in [[prd]] §7.2.

**body log** — A daily JSONL file (`body-YYYY-MM-DD.jsonl`) written to the configured log directory. Each line is one complete JSON record describing a single tapped request: metadata, captured request body, captured response body, truncation flags, and redaction flags. See [[prd]] §7.3 for the canonical schema.

**inference endpoint** — One of the four Ollama API paths the proxy is configured to body-log: `/api/generate`, `/api/chat`, `/api/embeddings`, `/api/embed`. Traffic to all other Ollama paths is proxied transparently but not body-logged. See [[prd]] §14.2.

**private upstream** — The Ollama server instance bound to `127.0.0.1:11435` rather than the public Ollama-compatible address. Moving Ollama to the private upstream port is the key topology change this project introduces. Clients continue to reach `0.0.0.0:11434`, which is now owned by the proxy. See [[prd]] §11.2.

**image redaction** — Log-only transformation that replaces any JSON field named `images` (case-insensitive, at any depth) with the literal string `"[redacted]"` before writing the body to disk. The actual traffic forwarded to Ollama is never altered. See [[prd]] §7.7.

**filename-based retention** — Retention strategy that deletes body log files based on the date embedded in their filename (`body-YYYY-MM-DD.jsonl`), not on file modification time. Files that do not match the naming pattern are ignored. Controlled by `OLLAMA_PROXY_RETENTION_DAYS`. See [[prd]] §14.3.

**body size cap** — The maximum number of bytes captured from a request or response body for logging purposes, controlled by `OLLAMA_PROXY_MAX_BODY_BYTES` (default 10 MiB). If the body exceeds the cap, the logged copy is truncated and the `request_truncated` or `response_truncated` flag is set to `true`. The actual traffic is never truncated. See [[prd]] §7.6.

**LaunchAgent** — A macOS `launchd` user-scope service definition stored in `~/Library/LaunchAgents`. This project uses two: `dev.ollama.server` (runs Ollama on the private upstream) and `dev.ollama.logging-proxy` (runs the proxy on the public port). See [[prd]] §11.3 and §11.4.

**MOC** (Map of Content) — An Obsidian note that functions as a curated index: links and one-line hooks, no body prose. `[[home]]` is this vault's master MOC. See [[vault-conventions]].

**ADR** (Architecture Decision Record) — A short, numbered, append-only document that records a single architectural or product choice, the context that made it live, the decision taken, and its consequences. ADRs live in `decisions/` and are never rewritten once accepted; to change a decision, a new ADR is written that supersedes the old one. See [[vault-conventions]].

---

## Related

- [[home]]
- [[vault-conventions]]