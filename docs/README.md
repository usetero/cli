# The Tero CLI Manual

This folder explains how to work in the CLI repository.

The code is still the source of truth for exact behavior. The job of this manual
is different. It should help someone understand the shape of the repository, the
main architectural boundaries, and the handful of patterns that matter enough to
be documented.

This repository is still in active development. New screens and services are
still being built. That makes it more important to document the stable mental
models instead of trying to catalog every current surface in detail.

## How To Read This

If you are new to the repository, read the foundations in order:

1. `foundations/01-what-this-repo-is.md`
2. `foundations/02-system-of-understanding.md`
3. `foundations/03-codebase-map.md`
4. `foundations/04-data-flow.md`
5. `foundations/05-hard-rules.md`

That sequence is meant to get you oriented quickly. It starts with what this
repository is responsible for, then explains the product model behind it, then
shows how that model appears in the code.

If you already understand the system and need implementation guidance, go to the
pattern docs.

## How The Docs Are Organized

- `foundations/` is the top-to-bottom introduction to the repository.
- `patterns/` captures recurring implementation rules that keep the codebase
  manageable.
- `meta/` defines how docs and `AGENTS.md` files should be maintained.

## What Belongs Here

This manual should explain:

- what the CLI owns and what it does not own,
- where truth lives,
- why local runtime and local data exist,
- how the major packages relate to one another,
- which patterns are important enough to defend explicitly.

It should not try to mirror the whole tree or become a second source of truth
for implementation details.
