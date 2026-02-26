# Logging

Use this doc when writing or changing code that emits logs.
For runtime troubleshooting workflows, see [OBSERVABILITY.md](OBSERVABILITY.md).

## Purpose

Logging in code should make behavior debuggable with low noise.

## Policy

1. Log state transitions and boundary events, not render loops.
2. Use structured key/value fields.
3. Use stable scope hierarchy via `scope.Child(...)` in constructors.
4. Keep debug logs cheap and remove one-off instrumentation after incidents.

## Levels

1. `Debug`: local diagnostics during development and targeted investigations.
2. `Info`: expected lifecycle milestones.
3. `Warn`: recoverable anomalies worth attention.
4. `Error`: behavior-impacting failures.

## Scope Rules

1. Create child scopes once in constructors.
2. Do not create ad-hoc nested scopes inside hot-path methods.
3. Add persistent context with `scope.With(...)` only when it should appear on all subsequent logs.
4. Use inline fields for one-off event context.

## Chat-Specific Guidance

1. Include identifiers (`turn_id`, `conversation_id`) on stream/tool lifecycle logs where available.
2. Log terminal stream outcomes and reasons (`completed`, `tool_use`, `aborted`, `failed`).
3. Avoid per-block/per-render diagnostics in steady state.

## Testing

Use `logtest.NewScope(t)` in tests to keep scoped logs attached to test output.
