# Data

This repo has two data movement patterns.

## 1) Query Path (CLI)

1. Parse user input.
2. Build API request.
3. Call control plane.
4. Render response.

Characteristics:

1. No local sync dependency.
2. Best for direct, stateless commands.

## 2) Sync Path (TUI)

1. Remote changes sync through PowerSync.
2. Data is materialized in local SQLite.
3. TUI reads local data for low-latency rendering.
4. Mutations are sent upstream; sync converges local state.

Characteristics:

1. Better UX for interactive views.
2. Works with transient network issues.

## Source of Truth

1. Control plane is authoritative.
2. SQLite is a read-optimized replica/cache.
3. Local projection bugs must never redefine business truth.

## Runtime Policy

1. Database operations should use `sqlite.WithTimeout(...)` unless a stricter deadline is already set.
2. Account scoping should prefer immutable scoped clients (`WithAccountID`) over mutating shared clients.
3. Chat query-tool results are intentionally capped (rows and bytes) to keep the TUI responsive.

## Chat Data Guarantees

1. Message history persists in SQLite.
2. Stream events are reduced before UI applies them.
3. Stream and tool events are turn-scoped.
4. User-cancelled partial assistant content is not persisted as committed assistant output.
5. Non-user aborts may be persisted with `stop_reason=aborted`.
