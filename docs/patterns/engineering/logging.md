# Logging

Logging in this repository should make the running system explainable.

The goal is not to produce more output. The goal is that when something goes
wrong in the CLI, an engineer can look at the logs and reconstruct what the app
thought was happening, which boundary owned the behavior, and where to go next.

This matters more in this repo than in a simple CLI because the app has
long-lived runtime behavior, local sync, and a TUI where failures can otherwise
look vague from the outside.

## Log Transitions And Boundaries

The most valuable logs in this repo are not generic debug chatter. They are the
logs that mark transitions and boundaries:

- onboarding state changes and failures,
- account runtime startup, shutdown, and readiness changes,
- syncer and uploader lifecycle changes,
- command failures,
- slow TUI update or render paths,
- any boundary where the app moves from one meaningful state to another.

Those are the logs that let someone reconstruct the story later.

## What To Log In Practice

Prioritize logs around:

- entering or leaving a meaningful step,
- starting or finishing a long-lived runtime component,
- retries, stalls, reconnects, or fatal errors,
- transitions from loading to ready or from ready to degraded,
- failures that a user can feel,
- places where the same surface can fail for multiple different reasons.

If a path is important to diagnose but impossible to reconstruct from logs, it
needs better instrumentation.

## Use Structured Fields, Not Narrative Blobs

Logs should be queryable.

That means using stable structured fields instead of relying on prose. Good
fields in this repo usually include:

- `organization_id`
- `account_id`
- current step or transition
- `msg_type`
- trigger or source
- `duration_ms`
- readiness or lifecycle state
- the concrete error

The message should still read clearly, but the fields should carry the durable
diagnostic value.

## Scope Logs By Ownership

Logs should come from the layer that owns the behavior.

This repo already uses scoped logging heavily through constructor-time
`scope.Child(...)`. Keep doing that. It matters because logs are much more
useful when they already tell you whether the event came from:

- a command surface,
- a TUI model,
- onboarding runtime,
- account runtime,
- PowerSync syncer or uploader,
- an infrastructure client.

If a log is useful only after you manually guess where it probably came from, it
is not scoped well enough.

## Severity Should Mean Something

Use levels consistently:

- `debug` for targeted diagnostic detail,
- `info` for expected lifecycle milestones and major transitions,
- `warn` for degraded or suspicious behavior that may recover,
- `error` for user-impacting failures or broken invariants.

If everything ends up at `info`, the log stops helping when a real problem
appears.

## High-Frequency Paths Need Discipline

Be especially careful in high-frequency paths such as TUI updates, renders, and
sync loops.

These paths should not emit constant chatter. If they need visibility, prefer
logs that fire only when something is actually noteworthy, such as a slow update
or a transition into an unexpected state.

This repo gets more value from a small number of high-signal runtime logs than
from a flood of low-value noise.

## Logs And User Feedback Are Different

Logs are for engineers. User-facing feedback is for the person running the app.

Do not try to make one channel do both jobs.

The useful split is:

- logs explain what happened and why,
- user-facing notices explain what the user needs to know or do next.

If a user sees an error but the logs cannot explain it, logging is incomplete.
If the logs are detailed but the user gets no understandable feedback, the
surface is incomplete.

## What Good Logging Looks Like By Area

### Onboarding

Onboarding logs should preserve transition context clearly enough that an
engineer can tell:

- what step the app believed it was on,
- what caused the transition,
- whether a remote action failed,
- whether the state was reloaded and reprojected.

### Account runtime

Account runtime logs should make lifecycle obvious:

- runtime started,
- runtime stopped,
- first sync completed,
- uploader or syncer errors occurred,
- readiness changed.

### TUI

TUI logs should be sparse and intentional.

The highest-value TUI logs are usually around slow paths, startup, shutdown, and
major surface failures. The TUI should not log every message just because it
can.

### Infrastructure and remote boundaries

Infrastructure logs should help distinguish:

- request/setup problems,
- auth failures,
- remote contract failures,
- transport or retry behavior,
- local storage corruption or invalid state.

The goal is to make the failure class clear before someone starts guessing.

## What Drift Looks Like

Logging is drifting when:

- important transitions are impossible to reconstruct,
- logs are noisy but still not useful,
- errors lack the identifiers or scope needed to explain them,
- user-visible failures have no corresponding engineering signal,
- high-frequency paths drown out meaningful state changes.

The fix is usually better instrumentation at the owning boundary, not more
generic logging everywhere.
