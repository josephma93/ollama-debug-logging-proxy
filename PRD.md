# PRD: Ollama Logging Proxy

## 1. Overview

This project introduces a lightweight HTTP reverse proxy that sits between local or LAN clients and a private Ollama server. Its purpose is to provide full request and response body observability for selected Ollama inference REST API calls while preserving compatibility with existing clients that already use the standard Ollama port.

The proxy will listen on the public Ollama-compatible address, forward traffic to a private Ollama instance bound to localhost on an internal port, and write structured logs containing request metadata, request bodies, and response bodies for tapped inference endpoints. The first version will focus on transparent behavior, reliable streaming support, image redaction, and bounded log retention.

## 2. Problem Statement

Ollama’s built-in logs provide useful runtime and access information, including endpoint, status code, source IP, timing, model loading, scheduler activity, and runner details. However, they do not expose the full JSON request body or response body. As a result, when a client calls the local Ollama REST API, the operator can see that a request occurred but cannot see the prompt, chat messages, model options, or generated response content.

This creates an observability gap when debugging local applications, LAN clients, scripts, automations, and services that call Ollama. The desired capability is to inspect exactly what was sent to Ollama and exactly what Ollama returned, without modifying Ollama itself or each individual client.

## 3. Goals

The proxy must provide transparent body-level observability for Ollama API traffic. It should preserve the existing client-facing address and port so current clients can continue using `http://<host>:11434` without code changes. Ollama itself should move behind the proxy and listen only on a private localhost address.

The first production-ready version should support streaming responses, because Ollama endpoints such as `/api/generate` and `/api/chat` commonly stream newline-delimited JSON chunks. The proxy must pass those chunks to clients as they arrive while also capturing a capped copy for logging.

The proxy must redact image payloads entirely from logs. Any JSON field named `images`, at any nesting depth, should be replaced before writing the request or response body to disk. This protects logs from becoming excessively large due to base64 image payloads and keeps the output readable.

Logs must be retained for a limited number of days. The default retention period is 10 days, controlled by an environment variable. For the intended deployment, the value should remain 10 days.

## 4. Non-Goals

The proxy is not intended to replace Ollama, modify model behavior, authenticate clients, or provide a full observability platform. It will not initially provide a web UI, search interface, metrics dashboard, OpenTelemetry integration, or long-term analytics storage.

The first version will not attempt to parse every Ollama response into normalized fields. Raw request and response bodies for tapped inference endpoints are the primary logging target. Ollama-aware extraction of fields such as `model`, `prompt`, `messages`, and reconstructed generated text may be added later.

The proxy will not solve privacy policy concerns for multi-user or production environments. For this deployment, the operator has explicitly accepted full prompt and response logging, with image redaction retained for log-size and readability reasons.

## 5. Target Environment

The initial deployment target is macOS using LaunchAgent services. Ollama is already installed as a macOS app and started through a custom LaunchAgent. The proxy should be compiled as a standalone local binary and started through its own LaunchAgent.

Default process layout:

```text
LAN or localhost clients
        ↓
Ollama Logging Proxy: 0.0.0.0:11434
        ↓
Private Ollama server: 127.0.0.1:11435
```

The proxy should follow normal macOS user LaunchAgent conventions for logs, using a default log directory under:

```text
~/Library/Logs/ollama-proxy
```

## 6. Users and Use Cases

### Primary user

The primary user is the machine owner/operator who runs local Ollama and wants to debug what clients are sending to the Ollama REST API.

### Main use cases

The operator wants to verify which application or host called Ollama, which endpoint was used, what prompt or chat messages were sent, what model response was returned, how long the request took, and whether requests are streaming successfully.

The operator also wants to keep existing client configuration unchanged. Any application already pointing to port `11434` should continue to work after the proxy is installed.

## 7. Functional Requirements

### 7.1 Transparent reverse proxy

The proxy must accept HTTP requests on a configurable listen address. The default listen address must be:

```text
0.0.0.0:11434
```

The proxy must forward all HTTP methods, paths, query strings, headers, and bodies to a configurable upstream Ollama target. The default target must be:

```text
http://127.0.0.1:11435
```

The proxy should not require Ollama-specific client changes. Clients should be able to keep using the usual Ollama base URL.

### 7.2 Normative body logging scope

This section defines the authoritative meaning of body-level observability, tapped requests, and body logs throughout this PRD. If any other section uses broader wording such as “for each proxied request,” “each request,” “Ollama API traffic,” “request body logging,” or “response body logging,” that wording must be interpreted according to this section.

The proxy must transparently proxy all Ollama API endpoints, including current and future endpoints, unless a proxy-owned endpoint is explicitly defined in this PRD. However, the proxy must only capture, redact, truncate, and store request and response bodies for the following tapped inference endpoints:

```text
/api/generate
/api/chat
/api/embeddings
/api/embed
```

A tapped request means a request whose normalized URL path exactly matches one of the tapped inference endpoints listed above. Query strings do not affect whether a request is tapped. For example, `/api/chat?x=1` is tapped because its path is `/api/chat`.

Requests to all other Ollama endpoints must still be proxied normally, but their request and response bodies must not be stored in body logs by default. Non-tapped requests may produce lightweight metadata-only service diagnostics if useful, but they must not write request bodies or response bodies to the daily body JSONL files.

All body logging requirements in this document apply only to tapped requests. This includes request body capture, response body capture, image redaction, body size caps, truncation flags, redaction indicators, daily body JSONL records, and inference logging validation.

### 7.3 Request metadata logging

For each tapped request, the proxy must write a structured JSONL body log entry containing at least:

```json
{
  "id": "unique request id",
  "started_at": "RFC3339 timestamp",
  "duration_ms": 1234,
  "client_ip": "192.168.1.11",
  "method": "POST",
  "path": "/api/generate",
  "query": "optional raw query string",
  "user_agent": "optional user agent",
  "status": 200
}
```

#### 7.3.1 Canonical MVP JSONL body log schema

This section defines the canonical JSONL record shape for the MVP. Each line in a daily `body-YYYY-MM-DD.jsonl` file must be one complete JSON object using this schema. Field order is not semantically significant, but implementations should keep this order when practical to make logs easier to read.

```json
{
  "id": "unique request id",
  "started_at": "RFC3339 timestamp",
  "duration_ms": 1234,
  "client_ip": "192.168.1.11",
  "method": "POST",
  "path": "/api/generate",
  "query": "optional raw query string",
  "user_agent": "optional user agent",
  "status": 200,
  "error": "",
  "request_body": "{\"model\":\"llama3\",\"prompt\":\"Say hello\",\"stream\":false}",
  "response_body": "{\"response\":\"Hello!\",\"done\":true}",
  "request_truncated": false,
  "response_truncated": false,
  "request_redacted": false,
  "response_redacted": false
}
```

Required field meanings:

| Field | Type | Meaning |
| --- | --- | --- |
| `id` | string | Unique request identifier generated by the proxy. |
| `started_at` | string | RFC3339 timestamp for when the proxy received the request. |
| `duration_ms` | number | Total proxied request duration in milliseconds. |
| `client_ip` | string | Remote client IP as observed by the proxy. |
| `method` | string | HTTP method sent by the client. |
| `path` | string | Normalized request path without query string. |
| `query` | string | Raw query string without the leading `?`; use an empty string when absent. |
| `user_agent` | string | Request user agent; use an empty string when absent. |
| `status` | number | HTTP status code returned to the client. |
| `error` | string | Proxy or upstream error message; use an empty string when no error occurred. |
| `request_body` | string | Captured request body after log-only redaction and truncation. |
| `response_body` | string | Captured response body after log-only redaction and truncation. |
| `request_truncated` | boolean | Whether the logged request body was truncated because of the capture limit. |
| `response_truncated` | boolean | Whether the logged response body was truncated because of the capture limit. |
| `request_redacted` | boolean | Whether at least one request body field was redacted before logging. |
| `response_redacted` | boolean | Whether at least one response body field was redacted before logging. |

For tapped requests without a request body, `request_body` must be an empty string and `request_truncated` must be `false`. For responses without a body, `response_body` must be an empty string and `response_truncated` must be `false`. If the upstream target is unavailable, the proxy must still write a JSONL entry for the tapped request with `status` set to `502`, `error` populated, and `response_body` containing the body returned to the client, if any.

### 7.4 Request body logging

For tapped requests, the proxy must capture and log the request body when one is present. The logged body should be stored as a string in the JSONL entry.

The proxy must not alter the actual request body sent to Ollama, except for normal proxy forwarding behavior. Redaction applies only to the log copy, not to traffic forwarded upstream.

### 7.5 Response body logging

For tapped requests, the proxy must capture and log the response body returned by Ollama. The logged body should be stored as a string in the JSONL entry.

For tapped streaming responses, the proxy must forward chunks to the client as they are received while simultaneously capturing a copy for the log. The proxy must not wait for the full response to finish before sending data to the client.

### 7.6 Body size cap

The proxy must enforce a configurable maximum captured body size for request and response logs. The default maximum should be:

```text
10485760 bytes
```

This value should be configurable through:

```text
OLLAMA_PROXY_MAX_BODY_BYTES
```

If the captured log body exceeds the cap, the proxy must truncate the logged copy and set the relevant truncation flag:

```json
{
  "request_truncated": true,
  "response_truncated": true
}
```

The cap must apply only to captured log data. It must not truncate the traffic sent to Ollama or the response returned to the client.

### 7.7 Image redaction

The proxy must redact JSON fields named `images` before writing request or response bodies to logs. Redaction must apply recursively at any depth.

Example input:

```json
{
  "model": "llava",
  "prompt": "Describe this image",
  "images": ["base64..."]
}
```

Logged output:

```json
{
  "model": "llava",
  "prompt": "Describe this image",
  "images": "[redacted]"
}
```

The field match should be case-insensitive. For example, `images`, `Images`, and `IMAGES` should all be redacted.

If the body is not valid JSON, the proxy should log it as-is and not fail the request.

### 7.8 JSONL body log files

The proxy must write body logs as JSON Lines files. The default log directory must be:

```text
~/Library/Logs/ollama-proxy
```

The log directory must be configurable through:

```text
OLLAMA_PROXY_LOG_DIR
```

The proxy must write daily log files using this naming format:

```text
body-YYYY-MM-DD.jsonl
```

Example:

```text
~/Library/Logs/ollama-proxy/body-2026-05-04.jsonl
```

Each line must represent a complete request/response record.

### 7.9 Log retention

The proxy must automatically delete body log files older than the configured retention period.

The default retention period must be 10 days:

```text
OLLAMA_PROXY_RETENTION_DAYS=10
```

For the intended deployment, the configured value should remain 10 days.

Retention should apply to files matching:

```text
body-*.jsonl
```

The proxy should run retention cleanup at startup and periodically while running. A reasonable first implementation is to run cleanup at startup and no more than once per hour during writes.

### 7.10 Service stdout and stderr logs

The LaunchAgent should send the proxy’s stdout and stderr to conventional service log files under the same log directory:

```text
~/Library/Logs/ollama-proxy/stdout.log
~/Library/Logs/ollama-proxy/stderr.log
```

These logs are separate from body logs. Body logs contain request/response data; stdout and stderr contain service runtime messages and errors.

### 7.11 Error handling

If the upstream Ollama target is unavailable, the proxy should return an HTTP `502 Bad Gateway` response and write a log entry with the error string.

If request body capture fails, the proxy should return an HTTP `400 Bad Request` response.

If writing to the body log fails, the proxy should report the error to stderr but should not fail the proxied request solely because logging failed.

## 8. Non-Functional Requirements

### 8.1 Performance

The proxy should add minimal latency. For streaming endpoints, the proxy must avoid buffering the full response before returning it to the client.

The log capture buffer must be bounded in memory according to the configured maximum body size. Large responses must continue streaming to clients even after the logged copy reaches the capture limit.

### 8.2 Reliability

The proxy should run continuously as a LaunchAgent with `KeepAlive` enabled. It should be able to restart automatically if it exits unexpectedly.

The proxy should create its log directory on startup if it does not already exist.

### 8.3 Compatibility

The proxy should support all Ollama API paths without needing explicit route definitions. It should forward unknown paths rather than rejecting them.

The proxy should preserve query strings, request methods, and request bodies. It should support normal Ollama clients that use `/api/tags`, `/api/generate`, `/api/chat`, `/api/embeddings`, and other current or future endpoints.

### 8.4 Operability

The proxy must be configurable through environment variables so LaunchAgent deployment remains simple. Configuration should not require a separate config file in the first version.

The proxy should emit basic startup information to stdout, including its listen address and upstream target.

## 9. Configuration

The proxy must support these environment variables:

| Variable                      |                       Default | Purpose                                        |
| ----------------------------- | ----------------------------: | ---------------------------------------------- |
| `OLLAMA_PROXY_LISTEN`         |               `0.0.0.0:11434` | Address and port where the proxy listens       |
| `OLLAMA_PROXY_TARGET`         |      `http://127.0.0.1:11435` | Private upstream Ollama URL                    |
| `OLLAMA_PROXY_LOG_DIR`        | `~/Library/Logs/ollama-proxy` | Directory for body logs and service logs       |
| `OLLAMA_PROXY_RETENTION_DAYS` |                          `10` | Number of days to keep body logs               |
| `OLLAMA_PROXY_MAX_BODY_BYTES` |                    `10485760` | Maximum captured request or response body size |

The Ollama LaunchAgent must set:

```text
OLLAMA_HOST=127.0.0.1:11435
```

The proxy LaunchAgent must set:

```text
OLLAMA_PROXY_LISTEN=0.0.0.0:11434
OLLAMA_PROXY_TARGET=http://127.0.0.1:11435
OLLAMA_PROXY_LOG_DIR=/Users/joseph/Library/Logs/ollama-proxy
OLLAMA_PROXY_RETENTION_DAYS=10
OLLAMA_PROXY_MAX_BODY_BYTES=10485760
```

## 10. CLI Requirements

The software should be distributed as a single binary that provides both the long-running proxy service and a small command-line interface for local operation and diagnostics.

The LaunchAgent should run the proxy through the `serve` subcommand:

```text
/Users/joseph/bin/ollama-logging-proxy serve
```

The CLI should use the same environment variables as the service where applicable, including `OLLAMA_PROXY_LISTEN`, `OLLAMA_PROXY_TARGET`, `OLLAMA_PROXY_LOG_DIR`, `OLLAMA_PROXY_RETENTION_DAYS`, and `OLLAMA_PROXY_MAX_BODY_BYTES`.

### 10.1 Initial CLI scope

The initial version should include these commands:

```text
ollama-logging-proxy serve
ollama-logging-proxy health
ollama-logging-proxy tail
ollama-logging-proxy purge
```

#### `serve`

Starts the reverse proxy service. This is the command used by LaunchAgent. It should bind to the configured listen address, forward inference traffic to the configured Ollama target, expose the proxy-owned health endpoint, write logs, redact images, and enforce filename-based retention.

#### `health`

Checks whether the proxy is running by calling the proxy-owned health endpoint:

```text
/__ollama_logging_proxy/health
```

By default, it should call the local proxy listener. It should print a simple human-readable result and exit with a non-zero status code if the proxy is unreachable or unhealthy.

#### `tail`

Reads recent entries from the current body log file. By default, it should read today’s log file from the configured log directory:

```text
body-YYYY-MM-DD.jsonl
```

A first implementation may print raw JSONL lines. Later versions may add pretty formatting, filtering by endpoint, filtering by client IP, filtering by request ID, and follow mode.

#### `purge`

Runs retention cleanup manually using filename-based retention. It should use the configured retention value from:

```text
OLLAMA_PROXY_RETENTION_DAYS
```

The default remains 10 days.

### 10.2 Future CLI commands

Future versions may add more advanced inspection commands:

```text
ollama-logging-proxy logs
ollama-logging-proxy inspect <request-id>
ollama-logging-proxy export
```

The `logs` command would list available body log files, dates, sizes, and optionally entry counts. The `inspect` command would print a readable version of a specific request by request ID. The `export` command would export selected entries into another format for analysis.

These commands are not required for the first version.

## 11. Expected Environment and System Configuration

This project expects a macOS host running Ollama and the Ollama Logging Proxy as separate user-level LaunchAgent services. A third party setting up the project should be able to reproduce the required environment from the details in this section without relying on prior context.

### 11.1 Required host environment

The initial supported environment is macOS. Ollama must be installed as the standard macOS application, with the Ollama server executable available at:

```text
/Applications/Ollama.app/Contents/Resources/ollama
```

The proxy must be installed as a standalone executable binary at a stable path. The recommended path is:

```text
/Users/joseph/bin/ollama-logging-proxy
```

The host must allow a user-level LaunchAgent to start both Ollama and the proxy. Both services should run under the same macOS user account that owns the local Ollama model directory and log directory.

The default Ollama model directory is expected to remain under the user account:

```text
/Users/joseph/.ollama/models
```

### 11.2 Required service topology

The system must run two separate services:

```text
com.joseph.ollama-server
com.joseph.ollama-proxy
```

The Ollama service must not listen directly on the public Ollama-compatible LAN port. Instead, Ollama must listen only on localhost using an internal upstream port:

```text
127.0.0.1:11435
```

The proxy must listen on the public Ollama-compatible address and port:

```text
0.0.0.0:11434
```

The proxy must forward traffic to Ollama at:

```text
http://127.0.0.1:11435
```

The resulting topology must be:

```text
Localhost or LAN clients
        ↓
ollama-logging-proxy listening on 0.0.0.0:11434
        ↓
Ollama listening privately on 127.0.0.1:11435
```

### 11.3 Ollama LaunchAgent requirement

A LaunchAgent named `com.joseph.ollama-server` must start Ollama using this executable and argument:

```text
/Applications/Ollama.app/Contents/Resources/ollama serve
```

The LaunchAgent must set this environment variable:

```text
OLLAMA_HOST=127.0.0.1:11435
```

This variable is required because it moves Ollama away from the default public port and makes it the private upstream service for the proxy.

The Ollama LaunchAgent should write its own stdout and stderr logs to a normal user log location, for example:

```text
/Users/joseph/Library/Logs/ollama/stdout.log
/Users/joseph/Library/Logs/ollama/stderr.log
```

Those logs are separate from the proxy body logs. Ollama logs are used for runtime diagnostics such as model loading, GPU discovery, scheduler activity, runner behavior, and process errors.

A representative Ollama LaunchAgent configuration should include:

```xml
<key>Label</key>
<string>com.joseph.ollama-server</string>

<key>ProgramArguments</key>
<array>
  <string>/Applications/Ollama.app/Contents/Resources/ollama</string>
  <string>serve</string>
</array>

<key>EnvironmentVariables</key>
<dict>
  <key>OLLAMA_HOST</key>
  <string>127.0.0.1:11435</string>
</dict>

<key>RunAtLoad</key>
<true/>

<key>KeepAlive</key>
<true/>
```

### 11.4 Proxy LaunchAgent requirement

A LaunchAgent named `com.joseph.ollama-proxy` must start the proxy using the `serve` subcommand:

```text
/Users/joseph/bin/ollama-logging-proxy serve
```

The proxy LaunchAgent must set these environment variables:

```text
OLLAMA_PROXY_LISTEN=0.0.0.0:11434
OLLAMA_PROXY_TARGET=http://127.0.0.1:11435
OLLAMA_PROXY_LOG_DIR=/Users/joseph/Library/Logs/ollama-proxy
OLLAMA_PROXY_RETENTION_DAYS=10
OLLAMA_PROXY_MAX_BODY_BYTES=10485760
```

The proxy LaunchAgent should write service stdout and stderr to:

```text
/Users/joseph/Library/Logs/ollama-proxy/stdout.log
/Users/joseph/Library/Logs/ollama-proxy/stderr.log
```

The proxy must write body logs to daily JSONL files in the configured log directory:

```text
/Users/joseph/Library/Logs/ollama-proxy/body-YYYY-MM-DD.jsonl
```

A representative proxy LaunchAgent configuration should include:

```xml
<key>Label</key>
<string>com.joseph.ollama-proxy</string>

<key>ProgramArguments</key>
<array>
  <string>/Users/joseph/bin/ollama-logging-proxy</string>
  <string>serve</string>
</array>

<key>EnvironmentVariables</key>
<dict>
  <key>OLLAMA_PROXY_LISTEN</key>
  <string>0.0.0.0:11434</string>

  <key>OLLAMA_PROXY_TARGET</key>
  <string>http://127.0.0.1:11435</string>

  <key>OLLAMA_PROXY_LOG_DIR</key>
  <string>/Users/joseph/Library/Logs/ollama-proxy</string>

  <key>OLLAMA_PROXY_RETENTION_DAYS</key>
  <string>10</string>

  <key>OLLAMA_PROXY_MAX_BODY_BYTES</key>
  <string>10485760</string>
</dict>

<key>RunAtLoad</key>
<true/>

<key>KeepAlive</key>
<true/>

<key>StandardOutPath</key>
<string>/Users/joseph/Library/Logs/ollama-proxy/stdout.log</string>

<key>StandardErrorPath</key>
<string>/Users/joseph/Library/Logs/ollama-proxy/stderr.log</string>
```

### 11.5 Port ownership requirement

Port ownership must be unambiguous.

Port `11434` must be owned by the proxy:

```text
0.0.0.0:11434 → ollama-logging-proxy
```

Port `11435` must be owned by Ollama:

```text
127.0.0.1:11435 → ollama serve
```

No Ollama process should remain bound to `0.0.0.0:11434` or `127.0.0.1:11434`. If Ollama is still bound to port `11434`, the proxy will not be able to start correctly on the public Ollama-compatible port.

A setup operator should verify the final state with:

```text
lsof -nP -iTCP:11434 -sTCP:LISTEN
lsof -nP -iTCP:11435 -sTCP:LISTEN
```

The first command should show the proxy process. The second command should show the Ollama process.

### 11.6 Client configuration requirement

Clients must use the proxy address, not the private Ollama upstream address.

Local clients should use:

```text
http://127.0.0.1:11434
```

LAN clients should use:

```text
http://<mac-lan-ip>:11434
```

Clients must not use:

```text
http://127.0.0.1:11435
```

Port `11435` is an internal upstream port reserved for proxy-to-Ollama traffic.

### 11.7 Firewall and network requirement

If LAN clients need access, the macOS host must allow inbound connections to the proxy on port `11434`.

The private Ollama upstream port `11435` should not be reachable from the LAN. Because Ollama is bound to `127.0.0.1`, only local processes on the macOS host should be able to connect to it.

The expected network posture is:

```text
LAN clients may reach Mac:11434
LAN clients may not reach Mac:11435
Local proxy may reach 127.0.0.1:11435
```

### 11.8 Health and validation requirement

The proxy must expose a proxy-owned health endpoint:

```text
/__ollama_logging_proxy/health
```

This endpoint must not be forwarded to Ollama.

A setup operator should validate the proxy with:

```text
curl http://127.0.0.1:11434/__ollama_logging_proxy/health
```

Expected result:

```json
{
  "ok": true,
  "service": "ollama-logging-proxy"
}
```

The operator should validate Ollama through the proxy with:

```text
curl http://127.0.0.1:11434/api/tags
```

The operator should validate the private upstream directly from the Mac with:

```text
curl http://127.0.0.1:11435/api/tags
```

A LAN client should be able to call:

```text
curl http://<mac-lan-ip>:11434/api/tags
```

A LAN client should not be able to call:

```text
curl http://<mac-lan-ip>:11435/api/tags
```

### 11.9 Inference logging requirement

The proxy must body-log inference endpoints only:

```text
/api/generate
/api/chat
/api/embeddings
/api/embed
```

Other Ollama endpoints should still be proxied, but their request and response bodies should not be stored in body logs by default.

For an inference request, a setup operator should be able to run:

```text
curl http://127.0.0.1:11434/api/generate \
  -d '{"model":"MODEL_NAME","prompt":"Say hello in one sentence.","stream":false}'
```

After the request completes, the operator should find a body log entry in:

```text
/Users/joseph/Library/Logs/ollama-proxy/body-YYYY-MM-DD.jsonl
```

The entry should include request metadata, the request body, the response body, truncation indicators, and redaction indicators.

## 12. Proposed Implementation

The first implementation should be written in Go and compiled into a standalone binary. Go is a good fit because it provides reliable HTTP reverse proxy support, low deployment overhead, straightforward streaming behavior, and simple distribution as a single executable.

The implementation should use Go’s standard HTTP reverse proxy primitives where possible. The response body should be wrapped with a custom `io.ReadCloser` that copies chunks into a bounded capture buffer while passing them through to the client.

Request bodies should be read into memory once, copied into a bounded capture buffer for logging, then restored so the upstream proxy can forward them to Ollama. Since Ollama request bodies are normally much smaller than model responses, this is acceptable for the first version. The body capture cap protects the log copy, but the complete request body must still be forwarded upstream.

Daily JSONL logging should be handled internally by the proxy. A mutex should protect writes so concurrent requests do not interleave log entries.

## 11. Deployment Plan

### 11.1 Prepare Ollama private port

Update the existing Ollama LaunchAgent so Ollama listens only on the private internal address:

```text
OLLAMA_HOST=127.0.0.1:11435
```

Reload the Ollama LaunchAgent after stopping any existing Ollama process bound to port `11434`.

### 11.2 Build the proxy

Compile the Go binary and place it at a stable path such as:

```text
/Users/joseph/bin/ollama-logging-proxy
```

The binary should be executable by the user account that owns the LaunchAgent.

### 11.3 Install proxy LaunchAgent

Create a LaunchAgent named:

```text
com.joseph.ollama-proxy
```

The LaunchAgent should start the proxy binary, set the required environment variables, enable `RunAtLoad`, enable `KeepAlive`, and redirect stdout/stderr to the proxy log directory.

### 11.4 Validate network behavior

After both services are running, validate that:

```text
lsof -nP -iTCP:11434 -sTCP:LISTEN
```

shows the proxy listening on port `11434`, and:

```text
lsof -nP -iTCP:11435 -sTCP:LISTEN
```

shows Ollama listening privately on port `11435`.

Then verify basic Ollama compatibility:

```text
curl http://127.0.0.1:11434/api/tags
```

Finally, issue a generate or chat request and verify that a new JSONL entry appears under:

```text
~/Library/Logs/ollama-proxy/body-YYYY-MM-DD.jsonl
```

## 12. Acceptance Criteria

The project is considered successful when all of the following are true:

1. Existing clients can call `http://<host>:11434` without changing their configuration.
2. Ollama itself is no longer exposed directly on `0.0.0.0:11434`.
3. The proxy forwards requests to `127.0.0.1:11435` successfully.
4. `/api/tags` returns normally through the proxy.
5. `/api/generate` returns normally through the proxy.
6. Streaming generation still streams to the client instead of waiting for the full response.
7. Each tapped request creates a JSONL body log entry containing metadata, request body, and response body.
8. JSON fields named `images` are replaced with `"[redacted]"` in logs.
9. Large logged bodies are truncated at the configured capture limit and marked as truncated.
10. Body logs are written to daily `body-YYYY-MM-DD.jsonl` files.
11. Body logs older than 10 days are removed automatically by default.
12. The log directory can be changed using `OLLAMA_PROXY_LOG_DIR`.
13. The retention window can be changed using `OLLAMA_PROXY_RETENTION_DAYS`.
14. Proxy stdout and stderr are written to separate service logs.

## 13. Future Enhancements

Later versions may add Ollama-aware parsing to extract normalized fields from raw bodies, including model name, prompt, system prompt, chat messages, model options, and reconstructed final generated text.

A future version may add a local inspection CLI for filtering logs by request ID, client IP, model, endpoint, or date. Another possible enhancement is a lightweight local web UI for reading recent requests and responses.

Additional redaction options may be added later, such as redacting specific JSON fields by name, replacing long strings, or disabling body logging for selected endpoints.

Metrics export may also be considered, including request counts, latency summaries, status code counts, token-related fields when available, and model usage summaries.

## 14. Product Decisions From Open Questions

### 14.1 Reconstructed generated text

The proxy should eventually reconstruct streaming model output into a separate convenience field while still preserving the raw response body. This reconstructed field must be clearly labeled as proxy-added derived data, not as a native Ollama response field.

Recommended future field name:

```json
{
  "proxy_generated_text": "Full reconstructed generated text from streaming chunks"
}
```

The `proxy_` prefix is required for derived fields added by this software. This naming convention makes it clear that the field is added by the logging proxy and was not returned directly by Ollama.

### 14.2 Endpoint body logging scope

The proxy must not body-log every Ollama endpoint. Body logging is normatively limited to endpoints normally used for inference, as defined in Section 7.2.

Initial inference endpoint allowlist:

```text
/api/generate
/api/chat
/api/embeddings
/api/embed
```

Requests to non-inference endpoints, such as `/api/tags`, should still be proxied normally. They may still produce lightweight metadata logs if useful, but they should not capture or store request and response bodies by default.

### 14.3 Filename-based retention

Log retention must be based on the date embedded in the filename, not file modification time.

Body log files must continue using this format:

```text
body-YYYY-MM-DD.jsonl
```

Retention cleanup should parse the `YYYY-MM-DD` portion of matching filenames and delete files whose filename date is older than the configured retention window. Files that do not match the expected filename pattern should be ignored by retention cleanup.

### 14.4 Proxy-owned health endpoint

The proxy should expose a simple health endpoint for manual diagnostics and service checks. The endpoint must be impossible to confuse with an Ollama endpoint.

Chosen health endpoint:

```text
/__ollama_logging_proxy/health
```

This endpoint is owned by the proxy and must not be forwarded to Ollama. It should return a simple JSON response such as:

```json
{
  "ok": true,
  "service": "ollama-logging-proxy"
}
```

The endpoint may later include upstream reachability, version, uptime, and log directory status, but the first version only needs a minimal success response.

### 14.5 Request headers

Request headers should be intentionally omitted from body logs in this version. The log should remain focused on request metadata, request body, response body, status, timing, truncation, and redaction indicators.

A future version may add optional header logging, but it is out of scope for the initial release.
