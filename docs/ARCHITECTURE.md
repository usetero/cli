# Architecture

Tero CLI is a presentation layer over the control plane.

## System Boundary

1. CLI does presentation, interaction, and transport orchestration.
2. Control plane owns business logic and authoritative state.
3. Local storage/sync exists for responsiveness, not ownership.

## Interfaces

1. TUI: interactive Bubble Tea interface under `internal/app`.
2. CLI commands: command entrypoints under `internal/cmd`.
3. MCP: planned transport surface documented in `MCP.md`.

## Dependency Rules

1. Domain models live in `internal/domain` and are shared.
2. Service code depends on interfaces, not concrete implementations.
3. Wiring/composition is done in `cmd/` only.
4. Presentation layers call services; services do not depend on presentation.

## Data Patterns

1. Direct query path: CLI command -> API client -> control plane.
2. Sync path: PowerSync -> SQLite -> TUI read models.

## TUI Message Contracts

For Bubble Tea in `internal/app`, messages are contracts and must be scoped.

1. Parent may broadcast messages to children, but children must only handle messages they own.
2. Periodic/timer messages must carry a source key (feature/tab/component) and handlers must filter by source.
3. Cross-feature messages live in `internal/app/msgs`.
4. Feature-local messages live in `internal/app/<feature>/msgs` (or unexported local types when strictly internal), except onboarding bootstrap contracts which live in `internal/core/bootstrap`.
5. Avoid generic shared message types without source identity.

## Chat Architecture

Chat is intentionally split into two layers:

1. `internal/chat`: stream protocol, reducer semantics, snapshot guarantees.
2. `internal/app/chat`: turn/round/message-list orchestration and rendering.

Key design choice:

1. Reducers determine transitions.
2. Handlers execute side effects.
3. Stream/tool events must be scoped to turn IDs.

## Directory Map

```text
cmd/                 composition and dependency wiring
internal/app/        Bubble Tea app/pages/components
internal/cmd/        CLI command handlers
internal/chat/       chat protocol/stream logic
internal/api/        control plane API client
internal/sqlite/     local database access
internal/powersync/  sync integration
internal/domain/     shared domain types
internal/auth/       authentication
```
