# Interfaces

These docs describe behavioral contracts for each user-facing surface.

Read these when you are changing interaction shape rather than internal
implementation details.

Each page explains what that surface promises to users, where responsibilities
start and stop, and which mistakes tend to create regressions.

## Pages

- [tui.md](tui.md): interactive runtime behavior and ownership boundaries.
- [cli.md](cli.md): command-surface adapter rules.
- [mcp.md](mcp.md): planned transport adapter constraints.

If a change affects more than one surface, start with
`../architecture/system-overview.md` first, then use these pages to verify each
surface still respects the same core boundaries.
