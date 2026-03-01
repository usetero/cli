# Message Contracts

This page defines the message naming and ownership contracts used by the Bubble
Tea runtime in `internal/app`.

These conventions are not style preferences. They are architectural guardrails
that keep update routing predictable and keep async behavior debuggable.

## Ownership boundaries

Use `internal/app/events` only for cross-feature shell messages that multiple
top-level app components consume.

Keep all other messages local to the feature package that owns the behavior.

If a message does not need to cross a top-level boundary, do not promote it to
`internal/app/events`.

## Naming conventions

### Cross-feature app events (`internal/app/events`)

Exported event struct names must end with one of:

- `Requested`
- `Published`
- `Changed`

Examples:

- `PaletteOpenRequested`
- `ErrorToastPublished`
- `SyncStateChanged`

### Feature-local async messages (`internal/app/**`)

Local message struct names should end in `Msg` and use semantic suffixes before
`Msg`, such as:

- `CompletedMsg`
- `LoadedMsg`
- `RequestedMsg`
- `TickMsg` / `PollTickMsg`
- `CreatedMsg`
- `ValidatedMsg`
- `RefreshedMsg`
- `UpdatedMsg`

This keeps logs/tests legible and prevents anonymous names like `detailMsg` or
`resultMsg` from spreading.

## Event-loop safety contract

`Update` and `View` are state/reduction code paths. They must not do blocking
or external work directly.

Blocking/external work belongs in `tea.Cmd` closures and should re-enter
`Update` through typed messages.

Examples of forbidden direct work in `Update`/`View`:

- network I/O,
- direct filesystem I/O,
- sleeps or command execution,
- other blocking calls.

## Enforced by lint

These contracts are enforced in `task lint` via:

- `scripts/lint/check-event-naming.sh`
- `scripts/lint/check-local-msg-naming.sh`
- `scripts/lint/check-event-loop-safety.sh`
- `scripts/lint/check-event-ownership.sh`
