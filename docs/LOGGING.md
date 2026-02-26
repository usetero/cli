# Logging

Logging should help debug behavior without becoming noise.

## Principles

1. Log decisions and state transitions, not every render path.
2. Scope logs by component (`scope.Child(...)`) so traces are readable.
3. Prefer structured fields over interpolated strings.
4. Keep debug logs cheap and remove one-off instrumentation after incidents.

## Levels

1. `Debug`: transient diagnostic detail.
2. `Info`: lifecycle milestones and expected transitions.
3. `Warn`: recoverable anomalies.
4. `Error`: failures requiring attention.

## Chat Logging Guidance

1. Log stream lifecycle boundaries and reasons (`completed`, `aborted`, `failed`).
2. Include turn identifiers for scoped events.
3. Avoid per-block rendering logs in normal operation.

## See Also

Operational and production observability notes live in [OBSERVABILITY.md](OBSERVABILITY.md).

