# Testing

Test behavior and invariants first. Coverage is a lagging indicator, not the goal.

## Test Types

1. Unit: deterministic, no network.
2. Integration: real external systems behind explicit tags/workflows.
3. Correctness/contract: validates assumptions at boundaries.

## Core Principles

1. Use real local dependencies when cheap (SQLite/theme/logger).
2. Mock only expensive or external boundaries.
3. Model async behavior with explicit messages and command draining.
4. Add a regression test for every production bug fix.
5. Add concurrency tests whenever scope/context is mutable or switchable at runtime.

## Chat Invariants (Must Stay Tested)

1. Event ordering: stream `seq` monotonic per turn.
2. Turn scoping: no cross-turn mutation from stream or tool events.
3. Terminal semantics: one terminal outcome per stream.
4. Cancellation semantics: `user_cancelled` is non-error behavior.
5. Persistence semantics: user-cancelled partial output is not committed assistant output.
6. Tool-loop integrity: no duplicate fire, no missing results.
7. History validity after cancel/retry/failure.

## Chat Test Layers

1. `internal/chat`: stream protocol and reducer correctness.
2. `internal/app/chat`: orchestration, persistence, lifecycle interactions.
3. `internal/app/chat/messagelist/behavior_test.go`: user-visible behavior scenarios.

## Minimum Chat Regression Matrix

1. `end_turn` success.
2. `tool_use` with all results.
3. results arriving before stream completion.
4. user cancel mid-stream.
5. non-user abort.
6. early stream failure.
7. rapid resubmit after cancel.
8. stale/out-of-order lifecycle or tool events ignored.

## Bubble Tea Testing Guidance

1. Drive components via `Update` messages.
2. Execute and inspect returned commands explicitly.
3. Keep `View` tests focused on rendering contracts.
4. Keep transition logic in pure reducers where possible.

## Client Scoping Tests

1. Cover `WithAccountID` behavior under rapid account switching.
2. Assert request headers use the intended account, not whichever account was set last.

## What Not To Test

1. Trivial getters/setters.
2. Framework internals.
3. One-line pass-throughs without policy.

## Required Command Before Merge (Chat Changes)

```bash
go test ./internal/chat ./internal/app/chat/... -count=1

## PowerSync Replay Fixtures

Use these commands for deterministic PowerSync correctness testing:

```bash
# 1) Capture raw fixture on demand (dev or prd)
TERO_ENV=dev task internal:powersync:capture OUTPUT=fixtures/powersync/dev-raw.ndjson DURATION=120s

# 2) Sanitize to commit-safe fixture
task internal:powersync:sanitize-fixture INPUT="$HOME/.tero/environments/dev/fixtures/powersync/dev-raw.ndjson" OUTPUT="internal/powersync/extension/testdata/dev-sanitized.ndjson" MAX_LINES=500

# 3) Run deterministic replay correctness test (uses committed fixture by default:
#    internal/powersync/extension/testdata/dev-sanitized.ndjson)
task test:correctness:powersync-replay

# 4) Run replay against a specific raw/sanitized fixture
task test:correctness:powersync-replay FIXTURE="$HOME/.tero/environments/prd/fixtures/powersync/prd-stream-2026-02-26.ndjson"
```

Rules:

1. Never commit raw production fixtures.
2. Commit only sanitized fixtures (prefer dev captures).
3. Keep committed fixtures small enough for fast CI.
```
