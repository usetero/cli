# Logging

Logging should help debug behavior quickly without spamming normal workflows.

## Principles

1. Log state transitions and boundary events.
2. Use structured fields, not interpolated strings.
3. Use scoped loggers (`scope.Child(...)`) for component provenance.
4. Remove temporary debug instrumentation after incidents are resolved.

## Levels

1. `Debug`: local diagnostic detail.
2. `Info`: expected lifecycle transitions.
3. `Warn`: unexpected but recoverable conditions.
4. `Error`: failures that impact behavior.

## Chat Logging Guidance

1. Log stream lifecycle boundaries and reasons.
2. Include `turn_id`/`conversation_id` when available.
3. Avoid per-render/per-block logging in steady state.

## Relationship to Observability

Use this document for logging policy and coding conventions.
Use [OBSERVABILITY.md](OBSERVABILITY.md) for operational diagnostics and runtime investigation patterns.
