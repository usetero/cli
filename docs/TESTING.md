# Testing

Test behavior and invariants, not fields and implementation details.

For chat specifically, treat the system as a state machine with async events. The highest-value tests prove we never enter invalid states under cancellation, tool fan-out, and streaming order edge cases.

## Test Types

**Unit tests**: fast, deterministic, no network.

```bash
task test
```

**Integration tests**: real services, credentials required.

```bash
task test:integration
```

**Correctness tests**: verify external contracts and assumptions.

```bash
task test:correctness
```

## Core Principles

1. Use real local dependencies (SQLite, theme, logger) whenever possible.
2. Mock only external or expensive dependencies.
3. Model async behavior with explicit messages, not sleeps.
4. Prefer scenario-based subtests over one-off assertions.
5. Every bug fix in chat flow adds a regression test.

## Chat Invariants

These invariants should be encoded in tests across `internal/chat` and `internal/app/chat`.

1. Event ordering:
`seq` is monotonic per turn; out-of-order events are rejected.
2. Turn scoping:
events and tool completions for one turn never mutate another turn.
3. Terminal semantics:
a stream ends in exactly one of `completed`, `tool_use`, `aborted`, or `failed`.
4. Cancellation semantics:
`user_cancelled` is not treated as an error.
5. Persistence policy:
user-cancelled partial assistant output is not persisted as a committed assistant turn.
Non-user aborts may be persisted with `stop_reason=aborted`.
6. Tool loop integrity:
no duplicate fire of tool results, no next-turn duplication, no missing results.
7. History validity:
message order/role alternation remains valid after cancel/retry/failure.

## Recommended Test Matrix

For any chat flow change, validate at least these scenarios:

1. `end_turn` success.
2. `tool_use` followed by all tool results.
3. tool results arriving before stream completion.
4. user cancel mid-stream.
5. non-user abort (context/network boundary).
6. stream failure before first assistant block.
7. rapid resubmit after cancel.
8. duplicate lifecycle message defense (`ToolResultsReady`, stream done).

## Bubble Tea Testing Strategy

Bubble Tea is testable if you separate reducers/state updates from rendering.

1. Drive state via message updates (`Update`) and drain commands explicitly.
2. Assert state transitions and emitted messages.
3. Keep `View()` tests focused on rendering contracts, not business logic.

Use helpers:

| Package | Purpose |
|---------|---------|
| `teatest` | Drain command loops; width and ANSI assertions |
| `logtest` | Scoped logger tied to `testing.T` |
| `dbtest` / `sqlitetest` | Real local DB fixtures |
| `chattest` | Mock chat client |

## Testing `View()`

Treat rendering as a contract:

1. Width never exceeded.
2. ANSI sequences remain valid.
3. Construction-time and post-resize paths both render correctly.
4. Parent-child render chains are tested with real models.

## What Not to Test

Skip low-value tests:

1. Trivial getters/setters.
2. Framework internals.
3. One-line delegators with no behavior.

## File Naming

| Type | File | Build Tag | Prefix |
|------|------|-----------|--------|
| Unit | `*_test.go` | none | `Test` |
| Integration | `*_integration_test.go` | `integration` | `TestIntegration_` |
| Correctness | `*_correctness_test.go` | `correctness` | `TestCorrectness_` |

## PR Checklist for Chat Changes

Before merging a chat-related change:

1. `internal/chat` reducer/client tests cover new lifecycle paths.
2. `internal/app/chat` tests cover orchestration and DB outcomes.
3. At least one regression test reproduces the issue fixed.
4. `go test ./internal/chat ./internal/app/chat/...` passes.
5. New invariants are documented here if behavior changed.
