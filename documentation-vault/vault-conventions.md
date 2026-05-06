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

## Related

- [[home]]
- [[glossary]]
- [Diátaxis framework](https://diataxis.fr/)