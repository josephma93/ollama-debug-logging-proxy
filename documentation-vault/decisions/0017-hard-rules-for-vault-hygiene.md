---
status: "Accepted"
---

# 0017 — Adopt Strict Hard Rules for Vault Hygiene

## Context

`vault-conventions.md` originally framed the documentation-vault structure as a set of *principles* — Diátaxis bucket placement, MOC discipline, append-only ADRs, link-density expectations, and a "what NOT to do" list. The principles are sound and the vault was rebuilt around them, but they are descriptive rather than enforceable. Five concrete stress points were identified in the macro-view discussion that produced this ADR:

1. Reintroduction of kitchen-sink notes under time pressure (the failure mode the original `index.md` exhibited).
2. `home.md` accumulating prose and reverting from a Map of Content into a content page.
3. Editing accepted ADRs in place rather than superseding them, which destroys the historical record.
4. Drift between code-rooted docs (`README.md`, `AGENTS.md`) and the vault, producing two contradictory sources of truth.
5. Hardcoded PII and machine-specific paths leaking into committed notes (currently visible in `reference/prd.md`).

Each of these is a foreseeable regression. Without a hard contract, the next contributor under deadline pressure will compromise on one of them and the structure will erode commit by commit.

## Decision

We will adopt a numbered set of strict, must-follow rules — the **Hard rules** block — and append it to `vault-conventions.md`. Each rule MUST use binding language (MUST, MUST NOT) and MUST ship with a check method, either an automatable shell predicate or an explicit "review-gated" flag. Rules with greppable checks are intended to be wired into a future `scripts/vault-lint.sh` runnable from CI; review-gated rules remain binding but caught at PR time.

The initial set covers seven categories: Structure, Note shape, ADRs and history, Linking, Source-of-truth boundaries, Hygiene, and Process. The exact list (R1–R22) is the source-of-truth in `vault-conventions.md`.

Adopting the rules is itself a change to the conventions, so per the existing pre-rule policy (and per the new R21), this ADR is the proposal that authorizes the edit. Future amendments to the hard-rules block follow the same loop.

## Consequences

- **Positive:** Each previously identified stress point now has a named rule and a check. The intent of the vault is no longer just a hope — it is a contract.
- **Positive:** The rule set is automation-ready. The greppable rules can be wired into CI without any rewording, and the review-gated rules are explicitly labeled so reviewers know what they are looking for.
- **Positive:** The hard-rules block is itself a worked example of the patterns the vault demonstrates: it was adopted via ADR (R21), it lives next to the conventions it amends, and any future revision must follow the same loop.
- **Trade-off:** Contributor friction goes up. A small documentation edit may now require thinking about whether it touches a rule. This is the explicit cost of strictness.
- **Trade-off:** Some rules (R15 two-hop reachability, R16 orphan detection) require tooling that does not yet exist. Until `scripts/vault-lint.sh` is built, those checks are aspirational and rely on reviewers running the documented commands manually.
- **Trade-off:** The rule set may itself need iteration. The first-cycle rules are best guesses; some will turn out to be too strict, too lax, or redundant once they meet real PRs. R21 makes evolving them a structured loop, not a free-for-all.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Pure social convention — keep `vault-conventions.md` descriptive only | Already failed in advance — the stress points listed above are exactly what social convention without hard rules produces under pressure. |
| Split the rules into a separate `vault-rules.md` file | Splits authority. Reviewers and contributors would have to consult two files; conventions and rules can drift apart. Keeping both in one file makes the relationship between principle and rule visible. |
| Skip rules, rely on linting tooling alone | Tooling can only check the automatable subset. Review-gated rules (R7, R10, R17, R18) cannot be expressed as scripts and require explicit written contracts. |
| Adopt fewer, broader rules (e.g., "keep the vault clean") | Vague rules are not enforceable. The whole point of strictness is that "kitchen-sink notes" is named and forbidden, not "messiness in general". |
| Use a third-party documentation linter (Vale, markdownlint) | These check prose style, not vault structure. They would not catch any of the five identified stress points. They may still be added later for prose-quality reasons, but they are orthogonal. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[home]]
- [[0015-launchagent-labels-com-joseph]]
- [[0016-launchagent-labels-dev-ollama]]
- [[prd]]
