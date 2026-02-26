# Architecture

Tero CLI is a presentation layer over the control plane. It exposes three interfaces:

1. TUI (`internal/app`, Bubble Tea models).
2. CLI commands (`internal/cmd`, direct API calls).
3. MCP server (`internal/mcp`, planned).

## Core Rules

1. The control plane is source of truth.
2. Local state exists for UX/perf only (cache/sync), not ownership.
3. Wiring lives in `cmd/`; domain and services stay implementation-agnostic.

## Data Patterns

1. Query path (CLI): command -> API client -> control plane.
2. Sync path (TUI/MCP): PowerSync -> local SQLite -> UI reads local projections.

## Chat Architecture

Chat has two layers with explicit boundaries:

1. `internal/chat`: stream protocol/reducer/client semantics.
2. `internal/app/chat`: Bubble Tea orchestration and rendering.

Design intent:

1. Stream events are reduced into stable snapshots.
2. Turn/round/message-list orchestration applies scoped events.
3. UI handlers execute side effects; reducers own transition policy.

## Directory Map

```text
cmd/                 composition/wiring
internal/app/        TUI pages/components
internal/cmd/        CLI commands
internal/chat/       chat streaming and protocol logic
internal/api/        control plane API client
internal/sqlite/     local storage
internal/powersync/  sync engine integration
internal/domain/     shared domain types
```

