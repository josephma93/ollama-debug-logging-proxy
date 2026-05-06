# Ollama Logging Proxy

A reverse proxy that sits in front of Ollama, preserves normal API behavior, and writes tapped request and response bodies to daily JSONL logs.

## Start here

- [[overview]] — what this project is and why it exists
- [[deployment-topology]] — the two-process model

## Explanation — the why

- [[overview]]
- [[deployment-topology]]
- [[execution-plan|Execution plan]] — how the project was sequenced

## Reference — the what

- [[prd|PRD]] — product requirements (source of truth for behavior)
- [[cli|CLI]] — `serve`, `health`, `tail`, `purge`
- [[launchd|launchd templates]] — service definitions and defaults
- [[scripts|Scripts]] — `install.sh`, `smoke-test.sh`, etc.
- [[release-model|Release model]] — Homebrew tap, stable vs canary

## How-to — the steps

- [[install|Install]]
- [[validate-setup|Validate the setup]]
- [[development-checks|Run local development checks]]

## History

- [[decisions/index|Decisions]] — append-only ADR record (18 ADRs)
- [[v0.1.0]] — first stable release after canary hardening
- [[releases/_template|Releases template]] — per-tag note template

## Related

- [[vault-conventions]] — how this vault is organized (Diátaxis + history + the R1–R22 hard rules)
- [[glossary]] — shared terms
