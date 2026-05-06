---
status: "Accepted"
---

# 0014 — Homebrew Tap with Stable + Canary Formulas, Hyphen-Suffix Tag = Prerelease

## Context

The README "Homebrew" section documents the supported install path: `brew tap josephma93/ollama-debug-logging-proxy` followed by `brew install`. Modern Homebrew requires formulae to come from a tap; local `--HEAD` installs from `./Formula/...` are not supported.

The release model (documented in `home.md` "Release Model" section and in the README) defines two channels:
- Stable: tags like `v0.1.0` — update the Homebrew formula
- Prerelease: tags like `v0.1.1-canary.1`, `v0.1.1-rc.1`, `v0.1.1-beta.1`, `v0.1.1-alpha.1` — any hyphen suffix marks a prerelease

Commit `63bbcb4` (2026-05-05) introduced the Homebrew prerelease flow with `.github/formula-template.rb` and `.github/workflows/release.yml`. Commit `0b913ec` split the formulas to support both stable and canary. Commit `b194456` restored the stable formula placeholder. Commit `520f92d` updated the formula for `v0.0.1`, the first stable release.

The GitHub Actions release workflow determines the channel from the git tag: a tag containing a hyphen after the semantic version core is a prerelease.

## Decision

We will distribute the proxy via a Homebrew tap with two formulas (stable and canary). The release channel is determined by the git tag: any hyphen suffix after the numeric core marks a prerelease. Stable tags update the stable formula; prerelease tags publish release artifacts without updating the stable formula.

## Consequences

- **Positive:** Users on the stable channel (`brew install`) get only production-quality releases. Canary adopters can opt into prerelease builds without affecting stable users.
- **Positive:** The release channel determination is rule-based (tag shape), not a manual workflow input. This eliminates human error in channel selection.
- **Positive:** Homebrew manages binary installation and PATH integration, matching macOS developer expectations.
- **Trade-off:** Maintaining a Homebrew tap requires keeping the formula's SHA256 hash updated on each release. The GitHub Actions workflow automates this, but a failed or misconfigured workflow leaves the tap stale. The sequence of commits `b194456` and `520f92d` shows this happened during initial setup.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Single Homebrew formula (no prerelease channel) | Would require stable-quality code for every published artifact. Canary testing becomes impossible without a separate mechanism. |
| GitHub Releases only (no Homebrew tap) | Users must manually download and install the binary. Homebrew is the standard macOS developer install mechanism and handles PATH, binary updates, and uninstall cleanly. |
| `brew install --HEAD` from repo | Modern Homebrew removed support for installing from a local `Formula/` directory outside a tap. README explicitly notes this is not supported. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[prd]]
- [[0003-user-launchagent-scope]]
- [[0005-single-binary-subcommands]]
