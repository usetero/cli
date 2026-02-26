# Documentation

How to write and maintain docs in this repository.

## The Goal

Docs should transfer understanding, not just facts. A reader should leave with a mental model and be able to reason about cases not explicitly listed.

## Principles

1. Start with why before how.
2. Teach invariants and decision boundaries.
3. Name tradeoffs and rejected alternatives when relevant.
4. Connect concepts to code paths.
5. Keep one source of truth per concept.

## What Not to Document

1. Language basics and obvious implementation details.
2. API signatures already visible in code/godoc.
3. Function-by-function walkthroughs with no architectural value.

## Maintenance Rules

1. Rewrite for coherence; do not append patchwork.
2. Delete stale content aggressively.
3. Update docs when behavior, invariants, or workflows change.
4. Prefer links over duplicated explanation.

## Docs in `docs/meta/`

1. `documentation.md` (this file): writing guidance for human docs.
2. `agent-docs.md`: writing guidance for `AGENTS.md` agent instruction files.

