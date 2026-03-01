# Architecture

These docs exist to help you choose the right boundary before you change code.
Most regressions in this repo do not come from bad syntax; they come from
correct local edits made in the wrong layer.

Read this section when you are:

- moving responsibilities between packages,
- changing cross-cutting runtime behavior,
- deciding where new logic should live.

## Suggested order

Start with [system-overview.md](system-overview.md), then read
[data-flow.md](data-flow.md), then [ui-architecture.md](ui-architecture.md).

That sequence mirrors how the system is built: first ownership boundaries, then
truth and data flow, then runtime execution behavior in the UI layer.

## How to use architecture docs with code

Use these pages to decide ownership first. Then open code under `cmd/`,
`internal/app/`, `internal/core/`, `internal/boundary/graphql/`, `internal/boundary/chat/`, `internal/boundary/powersync/`, and `internal/powersync/`
to implement inside that boundary.

The goal is not to memorize diagrams. The goal is to make design decisions that
stay consistent as the repository grows.
