# Documentation Standards

These docs should feel like a strong engineer helping another engineer build the
right model of the system. The goal is not documentation volume; the goal is
faster correct decisions.

If a reader finishes a page and still cannot predict how the system behaves
under failure or edge conditions, the page is not complete yet.

## What a good doc should do

A useful page in this repository should help a reader:

1. understand why this part of the system exists,
2. predict non-happy-path behavior,
3. identify what must not be broken,
4. find the right code entry points quickly.

If a page only lists files or commands with no model behind it, it is not done.

## Writing guidance

Start with context and intent, then explain runtime behavior, then anchor to
invariants and code paths. Use lists where they improve clarity, but avoid
defaulting to checklist style for conceptual docs.

Prefer prose for “read-to-understand” documents (architecture, domain).
Use compact lists for operational checklists and reference contracts.

## Maintenance guidance

Treat docs as part of the implementation contract:

1. update docs in the same change as behavior/workflow changes,
2. remove stale content aggressively,
3. keep one canonical page per concept and link to it instead of duplicating.

Coherence matters more than preserving old wording. Rewrite sections fully when
the model changes.
Incremental patching is fine for small corrections, but architectural shifts
usually deserve a deliberate rewrite so the mental model stays coherent.
