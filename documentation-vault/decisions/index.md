# Decisions — Map of Content

Architecture Decision Records (ADRs) capture the significant choices made while building `ollama-logging-proxy`, along with the context and tradeoffs that were live at the time each decision was made. ADRs are append-only: once accepted, a record is never rewritten. To change a decision, write a new ADR that supersedes the old one. See [[vault-conventions]] for the full rules governing this folder.

## Accepted

- [[0001-reverse-proxy-not-fork|0001 — Reverse Proxy in Front of Ollama, Don't Fork or Patch]] — establishes the foundational architectural approach: a transparent HTTP proxy in front of an unmodified Ollama process.
- [[0002-two-process-launchd-topology|0002 — Two-Process launchd Topology: Ollama Private 11435, Proxy Public 11434]] — splits Ollama and the proxy into separate services with unambiguous port ownership.
- [[0003-user-launchagent-scope|0003 — macOS User LaunchAgent (Not LaunchDaemon, Not brew services)]] — both services run as user-scope LaunchAgents, not system daemons or Homebrew-managed services.
- [[0004-go-stdlib-reverseproxy|0004 — Go + `net/http/httputil.ReverseProxy` from the Standard Library]] — the implementation language and core forwarding primitive, chosen for single-binary distribution and streaming correctness.
- [[0005-single-binary-subcommands|0005 — Single Binary Exposing `serve / health / tail / purge` Subcommands]] — one compiled binary handles both the long-running service and operator CLI commands.
- [[0006-env-var-only-config|0006 — Configuration via Environment Variables Only, No Config File]] — all configuration is via environment variables; no config file in the first version.
- [[0007-tapped-endpoint-allowlist|0007 — Body Logging Only for Tapped Inference Endpoints]] — body capture is limited to `/api/generate`, `/api/chat`, `/api/embeddings`, and `/api/embed`.
- [[0008-log-only-image-redaction|0008 — Image Redaction Applies Only to the Log Copy, Never to Upstream Traffic]] — the `images` field is redacted in logs but the original bytes are forwarded to Ollama intact.
- [[0009-filename-based-retention|0009 — Retention Parses the Date from `body-YYYY-MM-DD.jsonl`, Ignoring mtime]] — log file age is determined from the filename date, not filesystem metadata.
- [[0010-jsonl-daily-files|0010 — Body Logs Are Daily JSONL Files with Serialized (Mutex) Writes]] — one JSONL file per day, written with a mutex to prevent interleaved entries under concurrent requests.
- [[0011-proxy-owned-health-endpoint|0011 — Health Endpoint at `/__ollama_logging_proxy/health` Is Proxy-Owned and Never Forwarded]] — a dedicated proxy health path that cannot be confused with any Ollama API endpoint.
- [[0012-omit-headers-from-body-logs|0012 — Request Headers Are Intentionally Omitted from Body Log Records]] — the JSONL schema captures `user_agent` as a single field but excludes the full header map.
- [[0013-conventional-commits-gate|0013 — Conventional Commit Subjects Enforced via Local commit-msg Hook and CI]] — belt-and-suspenders enforcement of the Conventional Commit format at both local and CI layers.
- [[0014-homebrew-tap-stable-canary-channels|0014 — Homebrew Tap with Stable + Canary Formulas, Hyphen-Suffix Tag = Prerelease]] — two Homebrew formulas (stable and canary), with the release channel determined by git tag shape.
- [[0016-launchagent-labels-dev-ollama|0016 — Rename LaunchAgent Labels to `dev.ollama.logging-proxy` and `dev.ollama.server`]] — canonical service labels using the `dev.ollama.*` project namespace.
- [[0017-hard-rules-for-vault-hygiene|0017 — Adopt Strict Hard Rules for Vault Hygiene]] — converts the descriptive vault conventions into a numbered, enforceable contract (R1–R22) covering structure, note shape, ADR discipline, linking, source-of-truth boundaries, hygiene, and process.
- [[0018-remove-iteration-artifacts-from-contributor-guidance|0018 — Remove Iteration Artifacts from Vault Contributor Guidance]] — generalizes R1 to forbid any new top-level folder and strips warnings against invisible historical states from `vault-conventions.md` and `AGENTS.md`, keeping ADRs as the only place that names point-in-time artifacts.

## Superseded

- [[0015-launchagent-labels-com-joseph|0015 — Initial LaunchAgent Labels `com.joseph.ollama-proxy` and `com.joseph.ollama-server`]] — original personal-namespace labels from the initial code commit; superseded by [[0016-launchagent-labels-dev-ollama]].

## Related

- [[home]]
- [[vault-conventions]]
