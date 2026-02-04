# Architecture

The Tero CLI is a presentation layer for the Tero control plane. It never implements intelligence—no analyzing log patterns, calculating waste, or making decisions. When you're tempted to add logic that feels smart, stop. That belongs in the control plane.

## The Big Picture

```
                         Control Plane
            ┌────────────────┬────────────────┐
            │                │                │
            ▼                ▼                ▼
       Chat API        GraphQL API       PowerSync
      (streaming)        (CRUD)           (sync)
            │                │                │
            │                │                │
┌───────────┼────────────────┼────────────────┼───────────┐
│           │                │                │           │
│           │         ┌──────┴──────┐         │           │
│           │         │             │         ▼           │
│           │         │             │     SQLite          │
│           │         │             │   (local copy)      │
│           │         │             │         │           │
│           ▼         ▼             ▼         ▼           │
│          TUI       CLI           MCP       TUI/MCP      │
│       (streaming) (direct)    (direct)   (queries)     │
│                                                         │
│                       CLI Codebase                      │
└─────────────────────────────────────────────────────────┘
```

**Two data access patterns:**

1. **TUI and MCP** are long-lived. They start PowerSync, sync the catalog to local SQLite, then query locally. Fast, works offline, handles repeated browsing/filtering/searching.

2. **CLI** is short-lived. Each command hits the API directly and exits. Can't wait for sync—`tero services list` just calls GraphQL.

Both patterns share auth, domain types, and config. They differ in how they access data.

## Three Interfaces

**TUI** — Interactive terminal UI built on Bubbletea. Chat with the AI, browse the catalog. Syncs data locally for fast queries.

**MCP** — Model Context Protocol server for coding agents. Long-lived process, syncs data locally. Claude Code and similar agents connect via JSON-RPC over stdio.

**CLI** — Traditional commands for scripting. Short-lived, queries API directly. `tero services list`, `tero auth status`, etc.

## Code Structure

```
cmd/                    Wiring. Creates implementations, injects dependencies.
internal/
  tui/                  TUI. Bubbletea models, pages, components.
  cmd/                  CLI commands. Direct API access.
  chat/                 Chat client. Streaming, message accumulation.
  api/                  GraphQL client. Generated from schema.
  sqlite/               Local database. Schema, queries.
  powersync/            Sync engine. Keeps SQLite current.
  domain/               Shared types. Message, Block, IDs.
  auth/                 OAuth flow, token storage.
  config/               Configuration from env vars and files.
  upload/               Background queue for message persistence.
  log/                  Structured logging.
```

**Where does new code go?**

- TUI pages, components, chat UI → `internal/tui/`
- CLI commands (short-lived) → `internal/cmd/`
- Chat streaming and protocol → `internal/chat/`
- Control plane API client → `internal/api/`
- Shared domain types → `internal/domain/`
- Wiring and composition → `cmd/`

## Key Design Principle: Dependencies Point Inward

Services depend on interfaces, not implementations:

```go
// Auth service needs secure storage
type Service struct {
    storage SecureStorage  // interface, not concrete keyring
}

// Chat client needs auth for tokens
type client struct {
    auth Auth  // interface, not concrete auth service
}
```

This lets us:
- Test with mocks
- Swap implementations (OS keychain vs file storage)
- Keep coupling loose

**Wiring happens in `cmd/`**—that's the only place where concrete implementations get created and connected.

## What to Read Next

1. **[DATA.md](DATA.md)** — How data flows. Local vs remote, SQLite, PowerSync.

2. Then pick your interface:
   - **[TUI.md](TUI.md)** — Bubbletea patterns, model hierarchy, chat
   - **[MCP.md](MCP.md)** — Protocol, tools, how agents connect
   - **[CLI.md](CLI.md)** — Commands, direct API access

3. Supporting docs:
   - **[TESTING.md](TESTING.md)** — How to write tests
   - **[LOGGING.md](LOGGING.md)** — How to write logs
