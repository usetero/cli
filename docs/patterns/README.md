# Patterns

The foundation docs explain how the repository is structured.

The pattern docs explain how to build inside that structure without making it
harder to live with later.

This repository does not need a huge amount of doctrine. It does need a small
set of clear patterns in the places where drift creates real debt quickly. Those
are the topics documented here.

## What A Pattern Doc Should Do

A good pattern doc in this repo should answer questions like:

- how do we usually do this here,
- what shape should this code take,
- what should stay out of this layer,
- what kinds of drift are we trying to prevent.

If reading the page would not change how someone writes code, it probably is not
specific enough yet.

## Current Pattern Areas

Architecture:

- `architecture/tui.md`
- `architecture/read-models.md`
- `architecture/services.md`

Engineering:

- `engineering/code-shape.md`
- `engineering/logging.md`
- `engineering/testing.md`

## What Does Not Belong Here

If a page is mainly about understanding the repository for the first time, it
belongs in `foundations/`.

If a page is mainly a workflow for a recurring task, it belongs in
`playbooks/`.

These docs should stay practical and opinionated. Their job is to preserve the
small set of engineering choices that keep the codebase manageable.
