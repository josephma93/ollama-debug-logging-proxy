---
status: "Accepted"
---

# 0004 — Go + `net/http/httputil.ReverseProxy` from the Standard Library

## Context

PRD §12 states: "The first implementation should be written in Go and compiled into a standalone binary. Go is a good fit because it provides reliable HTTP reverse proxy support, low deployment overhead, straightforward streaming behavior, and simple distribution as a single executable."

The execution plan's "Recommended implementation approach" section reinforces this: "Use standard library `net/http/httputil.ReverseProxy` as the core" and "Minimize external dependencies (prefer standard library)." Commit `44527fa` (2026-05-04) implemented the proxy using `httputil.NewSingleHostReverseProxy`, reading from `go.mod` which declares Go 1.22.

The streaming requirement (PRD §7.5) was a key constraint: the proxy must forward response chunks to clients as they arrive without buffering the full response. Go's `httputil.ReverseProxy` provides this behavior with standard `http.Flusher` support, and the execution plan describes using `io.TeeReader` or custom `io.ReadCloser` wrappers for non-blocking body capture.

## Decision

We will implement the proxy in Go 1.22 using `net/http/httputil.ReverseProxy` from the standard library as the core forwarding mechanism, minimizing external dependencies.

## Consequences

- **Positive:** Single statically linked binary with no runtime dependencies. Trivial deployment: copy the binary, install the plist.
- **Positive:** `httputil.ReverseProxy` handles streaming, chunked transfer encoding, and WebSocket upgrades without custom implementation.
- **Positive:** Go's race detector (`go test -race`) is available for validating the concurrent log-write path (PRD §12 mentions the mutex requirement).
- **Trade-off:** Go is not the most ergonomic language for scripted log inspection or one-off tooling. The `tail` and `purge` subcommands are necessarily compiled, not interpretable scripts.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Rust + hyper | Strong performance and safety profile, but higher implementation complexity and slower compile times. No meaningful performance advantage for a single-user local proxy. |
| Node.js + `http-proxy` | Dynamic runtime, easy to iterate, but requires Node installed on the target macOS host. Produces a directory tree rather than a single binary. Streaming behavior requires careful async handling. |
| nginx + Lua filter | No Go compilation needed, but nginx is a heavyweight dependency for a personal debug tool, and Lua scripting for the JSONL logging and redaction logic adds significant operational complexity. |
| Python + WSGI proxy | Easy prototyping but poor streaming behavior with standard WSGI; requires asyncio or similar. Single-binary distribution is complex. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[execution-plan]]
- [[0005-single-binary-subcommands]]
- [[0010-jsonl-daily-files]]
