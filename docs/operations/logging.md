# Logging

Logging in this repository is a design tool, not just a debugging afterthought.
Good logs make runtime behavior explainable without turning the signal into
noise.

## What we log on purpose

We prioritize lifecycle and boundary events:

- onboarding gate transitions,
- runtime initialization/teardown boundaries,
- syncer/uploader lifecycle changes,
- stream/round terminal outcomes,
- explicit slow-loop warnings.

We do not log per-render chatter or high-frequency noise that cannot be acted on.

## Structured logging expectations

Use structured fields with stable keys. A log line should be queryable across
sessions, environments, and components.

Scopes should be created in constructors (`scope.Child(...)`) so all descendant
logs carry consistent ownership context.

When adding event fields, prefer durable identifiers and transition context over
incidental formatting details.
Good default fields are things like gate/account/context identifiers and
transition triggers, not prose fragments that are hard to query.

## Severity usage in this repo

- `debug`: useful during focused diagnosis, low consequence if missing.
- `info`: expected lifecycle milestones and major transitions.
- `warn`: recoverable anomalies or degraded behavior.
- `error`: user-impacting failures or broken invariants.

If everything is logged at `info`, nothing stands out during incident triage.

## Logging and user feedback are separate channels

A useful rule:

- logs explain system behavior to engineers,
- toasts explain actionable outcomes to users.

Do not rely on one channel to serve both audiences.
When in doubt, log the full diagnostic context and show a concise user-facing
toast with the actionable outcome.

## Onboarding-specific note

Onboarding logs should always preserve transition context (gate, trigger,
duration when applicable). This is essential for diagnosing “stuck” or
“jitter” reports in a deterministic flow engine.
