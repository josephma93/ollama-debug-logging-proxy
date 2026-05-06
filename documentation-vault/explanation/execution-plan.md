# Ollama Logging Proxy: Execution Plan

## Related

- [[home|Overview]]
- [[prd|PRD]]


## Purpose of this document

This document is not a replacement for the PRD and should not duplicate the PRD's product requirements. The PRD remains the source of truth for product scope, behavior, endpoint coverage, logging requirements, redaction rules, retention expectations, and acceptance criteria.

The goal of this document is different: it translates the PRD into an execution plan for getting the project off the ground. It focuses on engineering structure, delivery sequence, repository layout, quality automation, GitHub CI, and the first implementation milestones.

When there is any conflict between this document and the PRD, the PRD wins.

## Relationship to the PRD

The PRD should answer:

```text
What are we building?
Why are we building it?
What behavior is required?
What is in scope or out of scope?
How will we know the product works?
```

This execution plan should answer:

```text
How should we organize the codebase?
What should we build first?
How should we sequence the work?
What engineering quality gates should exist from day one?
How should local checks and GitHub CI support the project?
What is the practical path from empty repo to usable MVP?
```

This means the PRD should not be copied into this document. Instead, implementation tasks should point back to the PRD whenever they depend on product behavior.

## Product direction, in brief

The project should be implemented as a small infrastructure utility that places a transparent reverse proxy in front of Ollama. The proxy's product behavior, including listener addresses, upstream target, tapped endpoints, JSONL logging, body truncation, image redaction, retention, and service behavior, should be governed by the PRD.

This document assumes the PRD-defined product direction and focuses on how to build it in a maintainable way.

## Recommended implementation approach

The project should start as a Go-based single-binary service. Go is a practical fit because the proxy needs reliable HTTP handling, streaming-safe response forwarding, straightforward filesystem operations, strong tooling, simple static analysis, and low-friction local deployment.

Agents should follow these technical constraints:
- Use standard library `net/http/httputil.ReverseProxy` as the core.
- Minimize external dependencies (prefer standard library).
- Use `io.TeeReader` or custom `io.ReadCloser` wrappers for non-blocking body capture.
- Implement configuration via environment variables using a simple `internal/config` package.
- All logs should be written in JSONL format to `OLLAMA_PROXY_LOG_DIR`.

## Repository structure

A practical initial repository structure would be:

```text
ollama-logging-proxy/
├── .github/
│   └── workflows/
│       └── ci.yml
├── cmd/
│   └── ollama-proxy/
│       └── main.go
├── internal/
│   ├── config/      # Environment-based configuration
│   ├── proxy/       # Reverse proxy logic and body capture
│   ├── logging/     # JSONL file management and log record construction
│   ├── redact/      # Recursive JSON redaction
│   ├── retention/   # Filename-based log deletion
│   └── service/     # OS-specific service helpers
├── launchd/
│   ├── dev.ollama.logging-proxy.plist
│   └── dev.ollama.server.plist
├── scripts/
│   ├── install.sh
│   ├── uninstall.sh
│   └── smoke-test.sh
├── tests/           # Integration and streaming tests
├── .golangci.yml
├── go.mod
├── go.sum
├── justfile
└── README.md
```

## Milestone 1: Empty repo to working proxy skeleton

The first milestone should establish the repository, Go module, CLI entry point, basic reverse proxy behavior, `justfile`, linter configuration, and GitHub Actions workflow.

### Task 1.1: Initialize Project
- Run `go mod init github.com/joseph/ollama-logging-proxy`.
- Create the directory structure: `cmd/ollama-proxy`, `internal/config`, `internal/proxy`.
- Create a `README.md` with project description.

### Task 1.2: Local Quality Gate (Justfile)
- Create a `justfile` that includes `fmt`, `vet`, `lint`, `test`, and `check`.
- **Verify:** `just --list` shows all commands.

### Task 1.3: Basic Proxy Implementation
- Implement `internal/config` to read `OLLAMA_PROXY_LISTEN` and `OLLAMA_PROXY_TARGET`.
- Implement a minimal `internal/proxy` using `httputil.NewSingleHostReverseProxy`.
- Implement `main.go` to wire them together and start the server.
- **Verify:** Running the proxy and calling it with `curl` forwards the request to a mock upstream.

### Task 1.4: CI Setup
- Create `.github/workflows/ci.yml` to run `just check` on pull requests.
- **Verify:** Push to a branch (if possible) or verify the YAML syntax.

### Acceptance points:

```text
[ ] Go module is initialized.
[ ] CLI entry point starts the proxy on OLLAMA_PROXY_LISTEN.
[ ] Requests are forwarded to OLLAMA_PROXY_TARGET.
[ ] justfile exists and `just check` passes.
[ ] GitHub Actions workflow is present.
```

## Milestone 2: PRD-defined logging path

The second milestone should implement the logging path required by the PRD without expanding product scope beyond it.

### Task 2.1: Endpoint Classification
- Implement logic in `internal/proxy` to identify "tapped" endpoints (`/api/generate`, `/api/chat`, etc.) as per PRD 7.2.
- **Verify:** Unit tests for path matching (ignoring query strings).

### Task 2.2: Body Capture Mechanism
- Implement a bounded buffer capture in `internal/proxy`. 
- Use `httputil.ReverseProxy.GetBody` for requests and a custom `io.ReadCloser` for response bodies.
- Ensure the capture is "tapped" only for specific endpoints.
- **Verify:** Tests showing bodies are captured up to `OLLAMA_PROXY_MAX_BODY_BYTES`.

### Task 2.3: JSONL Logging
- Implement `internal/logging` to construct the record defined in PRD 7.3.1.
- Handle daily file rotation based on system time (`body-YYYY-MM-DD.jsonl`).
- **Verify:** Log files are created and contain valid JSONL.

### Acceptance points:

```text
[ ] PRD-defined tapped endpoints are classified correctly.
[ ] Request and response bodies are captured for tapped endpoints.
[ ] JSONL log records match the PRD schema.
[ ] Daily file naming follows `body-YYYY-MM-DD.jsonl`.
[ ] Non-tapped endpoints do not produce body logs.
```

## Milestone 3: Streaming correctness

The third milestone should focus on streaming behavior.

### Task 3.1: Incremental Forwarding
- Ensure response body capture does not buffer the entire response before sending to the client.
- **Verify:** Integration test with a slow streaming upstream server (e.g., sending chunks every 100ms) shows the client receives chunks immediately.

### Task 3.2: Race Condition Checks
- Run tests with `-race`.
- Ensure log writing (shared file/mutex) does not block the proxy path.

### Acceptance points:

```text
[ ] Streaming responses are forwarded incrementally.
[ ] Logging does not delay streamed chunks.
[ ] `go test -race ./...` passes.
```

## Milestone 4: Safety behavior from the PRD

The fourth milestone should implement and test redaction and truncation.

### Task 4.1: Recursive Redaction
- Implement `internal/redact` to find and replace `"images": [...]` with `"images": "[redacted]"` recursively.
- Match case-insensitively as per PRD 7.7.
- **Verify:** Comprehensive unit tests with nested and malformed JSON.

### Task 4.2: Truncation Logic
- If captured body exceeds `OLLAMA_PROXY_MAX_BODY_BYTES`, truncate the log string and set `request_truncated` or `response_truncated` to `true`.
- **Verify:** Tests with bodies larger than the limit.

### Acceptance points:

```text
[ ] PRD-defined redaction behavior is implemented and verified.
[ ] PRD-defined truncation behavior is implemented and verified.
[ ] Malformed JSON bodies are logged as-is without failing the request.
```

## Milestone 5: Local service operation

The fifth milestone should turn the working binary into a usable local service.

### Task 5.1: Retention Cleanup
- Implement `internal/retention` to delete `body-*.jsonl` files older than `OLLAMA_PROXY_RETENTION_DAYS` based on the filename date.
- Run cleanup at startup and once per hour.

### Task 5.2: LaunchAgent and Scripts
- Create `.plist` templates in `launchd/`.
- Create `scripts/install.sh` and `scripts/uninstall.sh` to handle directory creation, binary placement, and `launchctl` commands.

### Task 5.3: Smoke Test
- Create `scripts/smoke-test.sh` that starts the proxy, sends a request, and verifies the log file existence and content.

### Acceptance points:

```text
[ ] Retention logic removes old files correctly.
[ ] Install/uninstall scripts work on macOS.
[ ] Smoke test passes in a clean environment.
```

## Final readiness definition

The project is ready for an initial release when three things are true.

First, the implementation satisfies the PRD-defined MVP behavior. This document should not restate that behavior as an independent checklist; the PRD should be used for final product acceptance.

Second, the project is operationally usable as a local macOS service, with install, uninstall, service logs, body logs, and smoke testing in place.

Third, the engineering workflow is healthy: `just check` passes locally, GitHub Actions runs on pull requests, and GitHub CI blocks changes that fail formatting, linting, static analysis, tests, race checks where practical, security scanning, or configured maintainability rules.

## Non-goal of this document

This document should not become a second PRD. It should not maintain a duplicate list of product requirements, endpoint details, schema fields, default values, or product acceptance criteria. Those belong in the PRD.

This document should stay focused on execution: how to organize the project, how to sequence the work, and how to keep the codebase healthy while implementing the PRD.
