# Chat Core

Owns stream protocol semantics and snapshot reduction.

## Rules

1. Preserve event ordering guarantees (`seq` monotonic per turn).
2. Keep stream reduction logic pure and testable.
3. Expose terminal semantics explicitly (`completed`, `tool_use`, `aborted`, `failed`).
4. Do not leak transport quirks into app-layer orchestration.

## Implementation Guidance

1. Put transition policy in reducers.
2. Keep client streaming glue thin around reducers/snapshots.
3. Include turn/conversation identifiers in stream envelope fields where available.

## Required Tests for Changes

1. Reducer transition tests for new lifecycle paths.
2. Ordering/scoping regression tests.
3. Aborted/cancelled terminal behavior tests.

