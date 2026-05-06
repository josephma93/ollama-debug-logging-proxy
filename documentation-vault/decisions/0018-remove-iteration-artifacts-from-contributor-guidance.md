---
status: "Accepted"
---

# 0018 — Remove Iteration Artifacts from Vault Contributor Guidance

## Context

`vault-conventions.md` and the Documentation Vault section of `AGENTS.md` were authored during a rapid restructuring of the documentation vault. Several rules and bullet points encode warnings against the *specific* shapes encountered during that restructuring — for example, "no `projects/<name>/` wrapper" and "the pre-restructure `index.md` is the canonical bad example".

These warnings are useful only to a reader who saw the prior state. For a contributor who lands cold, they reference invisible history and burn attention without contributing rule-clarity. They function as **anti-rules**: they guard against shapes future contributors would not naturally produce while obscuring the more general principles underneath.

A clean line exists between two kinds of documents that handle history differently:

- **Conventions** (`vault-conventions.md`, the `AGENTS.md` Documentation Vault section) are forward-looking rules. They must read cleanly to a contributor who has no memory of how the project arrived at its current shape.
- **ADRs** (everything in `decisions/`) are point-in-time historical records. Naming artifacts of the moment they were written is exactly their job; R10 enforces that this naming is preserved.

This ADR cleans up the conventions side without touching the ADR side.

## Decision

Remove or generalize iteration artifacts in `vault-conventions.md` and `AGENTS.md`. Specifically:

- **R1** is rewritten from "no `projects/<name>/` wrapper" to a general rule: the vault top-level layout is fixed, and a new top-level folder requires an ADR amending these conventions. This catches a broader class of regressions and removes the artifact reference.
- The "What NOT to do" entry "Do not recreate `projects/<name>/`" is removed; the generalized R1 covers it.
- References to "the pre-restructure `index.md`" are removed from R7, the "What NOT to do" kitchen-sink entry, and the Mental-model intro paragraph. The principle ("notes must be single-intent") stands without the lost example.
- R19's parenthetical pointer ("`reference/prd.md` currently violates this; backfill is a known follow-up") is removed. Tracking the violation is the follow-up backlog's job, not the rule's.
- The `AGENTS.md` "Single-project root" bullet is removed. The bucket-placement bullet that follows already states the canonical buckets and root files, which implicitly forbids inventing new ones.

Existing ADRs are not touched. Iteration references inside ADRs (notably ADR 0017's Context) are kept as historical snapshots — interpretable by future readers as "an earlier state of the project", per the conventions/ADR distinction above.

## Consequences

- **Positive:** Contributor guidance reads as principles, not as the memory of one specific iteration. A new contributor learns *the rule*, not *our story*.
- **Positive:** R1 generalized to "no new top-level folders without an ADR" prevents a wider class of structural drift than the specific `projects/<name>/` warning did.
- **Positive:** Reinforces R21 by demonstrating that even small grooming passes go through the ADR loop. This ADR is the second non-supersession example after 0017.
- **Trade-off:** Some narrative texture about why the structure looks the way it does is lost from the conventions surface. That context now lives in git commit history and ADR 0017's Context section, both of which are still reachable but require deliberate digging.
- **Trade-off:** One more ADR in the chain. Marginal overhead.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Leave the artifacts in place | They will only get more confusing as the project ages and the original iteration drifts further out of memory. The `projects/<name>/` warning already requires explanation. |
| Move historical narrative into a separate "vault-history" note | The history is already in git commits and ADR 0017's Context. A dedicated narrative note duplicates that record and decays as the project evolves. |
| Keep the specific examples but add a "Background" section explaining them | Still forces future readers to consume narrative they don't need to follow rules. The point of conventions is that they read cleanly without context. |
| Edit ADR 0017 to remove its iteration references too | Forbidden by R10 (ADRs are append-only); also wrong on principle — ADRs are supposed to name the state they reasoned about. |

## Supersedes

## Superseded by

## Related

- [[vault-conventions]]
- [[0017-hard-rules-for-vault-hygiene]]
- [[home]]
