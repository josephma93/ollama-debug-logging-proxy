# Vault Conventions

This vault is the project's documentation-of-record. Read this before adding, moving, splitting, or deleting notes. Keeping the structure consistent is what makes the vault feel like a network of interconnected knowledge instead of a graveyard of orphan notes.

## Mental model: Diátaxis + history

The vault is organized using the [Diátaxis](https://diataxis.fr/) framework, augmented with a history layer for project-specific decisions and releases. Every note belongs in exactly one bucket. If you can't decide which bucket fits, the note is too broad — split it.

```
documentation-vault/
├── home.md                ← master MOC; the front door
├── vault-conventions.md   ← this file
├── glossary.md            ← shared terms
├── explanation/           ← the WHY
├── reference/             ← the WHAT
├── how-to/                ← the HOW
├── decisions/             ← ADRs (history)
└── releases/              ← per-tag notes (history)
```

This vault is single-project. There is no `projects/<name>/` wrapper, and adding one is a regression — the root *is* the project.

## The four documentation buckets

### explanation/ — the WHY

Discussion-oriented. Reading these builds understanding, not action.

- Architecture, design rationale, system topology, mental models, project background.
- Examples: "Topology overview", "Why streaming matters", "Why image redaction is log-only".
- If a note tells you *what to do*, it belongs in `how-to/`. If it tells you *what is true*, it belongs in `reference/`.

### reference/ — the WHAT

Information-oriented. Reading these surfaces facts. No tutorials, no rationale.

- PRD, env vars, JSONL schema, CLI surface, plist schema.
- Dry, accurate, scannable. If you find paragraphs explaining *why*, move them to `explanation/` and link.

### how-to/ — the HOW

Task-oriented. Reading these gets a single job done.

- Examples: "Install on macOS", "Run a smoke test", "Recover from a port conflict", "Cut a release".
- One note, one goal. No theory dumps; link to `explanation/` for context.

## The history layer

### decisions/ — the record of choices

Architecture Decision Records (ADRs), append-only.

- One file per decision, numbered: `0001-redact-images.md`, `0002-filename-based-retention.md`.
- Start from `decisions/_template.md`.
- Status flows: Proposed → Accepted → Superseded by [[0007-...]].
- Never delete or rewrite an accepted ADR. To change a decision, write a new ADR that supersedes it.
- The ADR captures the *tradeoff that was live at the time*. That historical view is the whole point.

### releases/ — what shipped when

One note per git tag.

- File names mirror tags: `v0.0.1.md`, `v0.1.0.md`.
- Contents: summary, highlights, breaking changes, migration notes, links to relevant ADRs.
- Start from `releases/_template.md`.

## Linking is more important than placement

Folders are shelves. Links are the graph. The "network of interconnected knowledge" property only emerges if every note participates in the link graph.

- Every note ends with a `## Related` section listing wikilinks to notes a reader is likely to want next.
- `home.md` is a Map of Content (MOC) — links and one-line hooks, no body prose.
- Each top-level folder may also grow its own MOC (`explanation/index.md`, etc.) once it exceeds ~5 notes.
- Prefer short wikilinks when filenames are unique: `[[prd]]` over `[[reference/prd|prd]]`. Disambiguate only when needed.
- Design for backlinks. The reference note describing the JSONL schema should be linked from every how-to and explanation that touches logging — that density is what readers experience as "the vault knows things".

## Naming

- Filenames are kebab-case: `execution-plan.md`, `port-conflict-recovery.md`.
- ADRs are zero-padded and start with the subject of the decision: `0007-stream-chunk-buffering.md`.
- No dates in filenames except releases (`v0.0.1.md`) and the ADR numeric prefix.
- Folders are lowercase singular nouns matching the four buckets above.

## What NOT to do

- **Do not recreate `projects/<name>/`.** This vault is single-project.
- **Do not duplicate code-rooted docs.** The repo has `README.md`, `AGENTS.md`, and inline comments — those describe the *current state* of the code. The vault explains the *thinking*. If a fact is true today only because someone made a choice, that choice belongs in `decisions/`. If a fact is a stable lookup, it belongs in `reference/`.
- **Do not put PII or machine-specific paths in committed notes.** Use `$HOME`, `<user>`, or generic placeholders. Hardcoded paths like `/Users/<name>/...` rot the moment another contributor reads the vault.
- **Do not let notes grow into kitchen-sink pages.** If a note spans more than one Diátaxis bucket, split it. The pre-restructure `index.md` mixed overview, install runbook, and CLI reference — that is the failure mode this layout exists to prevent.
- **Do not leave notes orphaned.** Every note must be reachable from `home.md` in at most two hops, and must link to at least one other note.

## Adding a new note: decision flow

1. Is it answering "how do I X?" → `how-to/`.
2. Is it a fact someone needs to look up? → `reference/`.
3. Is it explaining why something is the way it is? → `explanation/`.
4. Is it a record of a choice that locked in a tradeoff? → `decisions/` (new numbered ADR).
5. Is it the changelog for a tagged version? → `releases/`.
6. None of the above? → it probably doesn't belong in this vault. Consider whether it belongs in the repo's `README.md`, `AGENTS.md`, or inline code comments instead.

## Updating these conventions

These conventions are themselves a decision. If they need to change, write an ADR proposing the change, link it from this file, and only then edit this file. The ADR records the *why*; this file records the *current rule*.

## Hard rules

These rules are non-negotiable. The pedagogical sections above describe the *target shape*; humans drift under time pressure, so the rules below close the gap. Each rule uses MUST or MUST NOT — never "should" — and ships with a check method. Where the check is a shell-greppable predicate, it is the canonical enforcement and will eventually live in `scripts/vault-lint.sh`. Where the check is "review-gated", the rule is still binding but caught by human review at PR time.

Violation of any rule blocks merge.

The hard-rules block was adopted via [[0017-hard-rules-for-vault-hygiene]]. Future changes to it follow R21.

### Structure

**R1.** The vault MUST have a single project root. No `projects/<name>/` wrapper, ever.
*Check:* `find documentation-vault -maxdepth 1 -type d` returns only the five buckets, `.obsidian`, and the vault root itself.

**R2.** The vault root is reserved. Only `home.md`, `vault-conventions.md`, and `glossary.md` may sit at the root. Every other note MUST live in a bucket.
*Check:* `find documentation-vault -maxdepth 1 -name '*.md'` lists exactly those three files.

**R3.** Every note MUST live in exactly one bucket: `explanation/`, `reference/`, `how-to/`, `decisions/`, or `releases/`.
*Check:* No `*.md` file under `documentation-vault/` outside those folders (root meta files excepted).

**R4.** Filenames MUST match `[a-z0-9_-]+\.md`. ADRs MUST match `\d{4}-[a-z0-9-]+\.md`. Releases MUST mirror git tags exactly (`v0.0.1.md`, `v0.1.0-canary.1.md`). Templates start with `_`.
*Check:* `find` + regex.

### Note shape

**R5.** Every note MUST start with an H1 title.
*Check:* The first non-frontmatter, non-blank line of every note begins with `# `.

**R6.** Every note MUST end with a `## Related` section containing at least one wikilink. MOCs and templates included.
*Check:* `grep -L '^## Related' documentation-vault/**/*.md` returns nothing.

**R7.** A note MUST cover exactly one Diátaxis intent (why / what / how). If it spans two, split it. The pre-restructure `index.md` is the canonical bad example.
*Check:* Review-gated.

**R8.** MOCs (`home.md`, `decisions/index.md`, any future map of content) MUST contain only: H1, an intro of ≤ 2 sentences, grouped wikilink lists with one-line hooks, and `## Related`. No body prose elsewhere.
*Check:* Review-gated.

### ADRs and history

**R9.** ADRs MUST follow `_template.md` exactly: frontmatter `status` plus the sections `## Context`, `## Decision`, `## Consequences`, `## Alternatives considered`, `## Supersedes`, `## Superseded by`, `## Related`, in that order.
*Check:* Structural lint script.

**R10.** Accepted ADRs are append-only. The only allowed edits are (a) status flip to `"Superseded"` with `## Superseded by` filled, or (b) typo / dead-link repairs that do not change meaning. To change a decision, write a new ADR and supersede.
*Check:* Review-gated.

**R11.** Any commit that adds or changes a `decisions/NNNN-*.md` MUST also update `decisions/index.md` in the same commit.
*Check:* CI script — if a `decisions/` ADR file is in the diff, `decisions/index.md` must also be in the diff.

**R12.** Every git tag matching `^v\d+\.\d+\.\d+(-.+)?$` MUST have a corresponding `releases/<tag>.md` note before the tag is pushed.
*Check:* For every tag `T` on origin, `releases/T.md` exists.

### Linking

**R13.** Wikilinks MUST use the short form `[[name]]`, not `[[bucket/name]]`. Allowed path-qualified exceptions: `[[decisions/index]]`, `[[releases/_template]]`.
*Check:* `grep -rEn '\[\[[a-z]+/' documentation-vault --include='*.md' --exclude=vault-conventions.md | grep -vE '\[\[(decisions/index|releases/_template)(\||\])'` returns nothing. The check excludes `vault-conventions.md` because this file legitimately quotes the forbidden pattern as part of teaching the rule.

**R14.** Every wikilink MUST resolve to a real file. Broken links block merge — either create the target or remove the link.
*Check:* `obsidian-cli deadends` plus a custom resolver script.

**R15.** Every note MUST be reachable from `home.md` in ≤ 2 wikilink hops.
*Check:* Graph traversal in `vault-lint.sh` (future).

**R16.** No orphan notes. Every note MUST have at least one incoming wikilink from a note that is not itself.
*Check:* `obsidian-cli orphans` returns no markdown files.

### Source-of-truth boundaries

**R17.** `README.md` and `AGENTS.md` describe current code state. The vault explains thinking. A fact MUST live in exactly one of those two spheres; the other links to it. Verbatim duplication is forbidden.
*Check:* Review-gated.

**R18.** PRD-defined behavior is owned by `reference/prd.md`. Other reference notes summarize and link; they MUST NOT restate PRD requirements verbatim.
*Check:* Review-gated.

### Hygiene

**R19.** No hardcoded PII or machine paths in committed notes. Use `$HOME`, `<user>`, `<host>`, `<lan-ip>`. (`reference/prd.md` currently violates this; backfill is a known follow-up.)
*Check:* `grep -rE '/Users/[a-z.]+|192\.168\.[0-9]+\.[0-9]+' documentation-vault/**/*.md` returns nothing.

**R20.** No `TODO`, `FIXME`, or `XXX` markers in committed notes. If a note is not ready, it does not merge. Follow-ups go in tracked issues, not the prose.
*Check:* `grep -rEn '\b(TODO|FIXME|XXX)\b' documentation-vault --include='*.md' --exclude=vault-conventions.md` returns nothing. The check excludes `vault-conventions.md` because this file legitimately names the forbidden markers in stating the rule.

### Process

**R21.** Changes to `vault-conventions.md` (including this Hard rules block) MUST come with a new ADR proposing the change, linked from the conventions file, in the same commit.
*Check:* Review-gated.

**R22.** Renaming or moving a linked note MUST use `obsidian-cli move` or an equivalent that atomically rewrites wikilinks across the vault. Plain `mv` on linked notes is forbidden.
*Check:* Review-gated — look for orphaned wikilinks in the PR diff.

## Enforcement

Rules with grep-able or scriptable checks should be wired into `scripts/vault-lint.sh` and run from CI alongside `just check`. Until that script exists, automatable rules are still binding — they're enforced manually by reviewers running the documented check command.

First-cycle automation priority (cheap to script, high false-positive resistance): R5, R6, R8, R11, R13, R14, R20.

Permanently review-gated (require human judgment): R7, R10, R17, R18, R21, R22.

Conventions changes go through ADRs (R21). The hard-rules block itself was adopted via [[0017-hard-rules-for-vault-hygiene]] — that ADR is the worked example.

## Related

- [[home]]
- [[glossary]]
- [[0017-hard-rules-for-vault-hygiene]]
- [Diátaxis framework](https://diataxis.fr/)