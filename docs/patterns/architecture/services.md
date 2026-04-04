# Services

Service boundaries in this repository should stay aligned with business
entities, not transport details.

That matters because the CLI has more than one valid way to interact with the
same conceptual data. Some flows need authoritative control-plane state. Others
need fast local reads over synced SQLite state. The service pattern here has to
support both without becoming confusing.

## Align Services With The Entity

The stable part of the design is the entity boundary.

If the repo has a tenancy concept, a catalog concept, an integration concept, or
some other business-shaped concept, the service boundary should reflect that
concept. Whether the implementation is local or remote is secondary.

That is why the useful pattern is usually:

- one service boundary per domain concept,
- local and remote implementations when the product genuinely needs both.

## Why Local And Remote Implementations Exist

The control plane is authoritative, but the CLI also needs a fast local runtime
once synced account-scoped state exists.

Because of that, some services legitimately have two implementations:

- a remote path against the control plane,
- a local path against synced SQLite data.

The tenancy and catalog packages already show this pattern.

The useful question is not "why are there two versions?" The useful question is
"which path should this flow be using right now?"

## When To Use Remote Services

Remote services are the right choice when the flow needs control-plane truth.

That is especially common when:

- the data is not expected to exist locally yet,
- the app is still in bootstrap,
- the operation is naturally a remote mutation,
- the flow is a direct command that does not need the local runtime.

## When To Use Local Services

Local services are the right choice when the data is already synced and the
surface benefits from low-latency local access.

That is especially common in the steady-state TUI, where repeated remote calls
would make the experience feel unnecessarily slow or shallow.

## Keep Queries Close To The Owner

The local side of this pattern often comes with package-local SQL and sqlc code.

That is not a smell in this repo. It is often the clearest way to support a
specific local read without forcing every query through one generic data layer.

If a query exists to support one service, keep it with that service.

## What Drift Looks Like

The service pattern is drifting when:

- local services start being treated like authority,
- remote services get bypassed for writes that should stay remote,
- the API shape follows transport details instead of the entity,
- callers cannot tell which path they should use for a given flow.

The goal is not to hide the existence of multiple paths. The goal is to keep the
choice explicit and the boundary easy to follow.
