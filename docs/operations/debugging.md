# Debugging

Use this doc when you are diagnosing behavior in the running CLI.

The main idea is simple: use repeatable, task-driven workflows in safe
environments first, then narrow scope with logs and targeted traces.

## Start with environment discipline

Debug workflows default to `dev` in `taskfiles/debug.yml`.
That is intentional. Iterative debugging in `prd` should be rare and explicit.

The debug taskfile includes a guard that blocks accidental `prd` use unless
`ALLOW_PRD=1` is set. Keep that safeguard.

## A practical triage flow

When something “feels broken” in the UI, start here:

1. Run through the debug task entrypoints.
2. Tail slow-loop logs and feature-specific traces.
3. Reproduce once with minimal noise.
4. Correlate logs with expected state transitions.
5. Turn findings into a test or a targeted fix.

The important part is reproducibility. Manual ad hoc commands tend to drift and
make incident analysis harder to compare across engineers.

## Core workflows you will use most

```bash
task debug:list
task debug:ui:run
task debug:ui:logs:slow
task debug:onboarding:trace
task debug:sync:trace
task debug:sync:health
task debug:data:snapshot
```

These commands are designed to align with the architecture: UI loop behavior,
onboarding transitions, sync/upload health, and local DB state.

## Diagnosing input lag or “frozen” UI

Use `task debug:ui:logs:slow` while reproducing.
The app emits `slow app update` and `slow app render` telemetry with fields
like `duration_ms`, `msg_type`, overlay state, and drawer context.

That gives you immediate signal on whether the event loop is blocked in update,
render, or a specific interaction mode.

## Diagnosing onboarding stalls or jitter

Use `task debug:onboarding:trace`.
Correlate gate enter/exit logs (including trigger and duration) with what the
user reports seeing. This tells you whether the issue is true stalling, rapid
gate transitions, or missing visibility/status feedback.

## Diagnosing sync/upload problems

Use `task debug:sync:trace` and `task debug:sync:health`.
These views combine syncer and uploader signals so you can distinguish
connectivity/retry behavior from queue stalls and normal post-failure
convergence.

## Fixture-driven sync debugging

For deeper PowerSync analysis, capture and sanitize fixtures:

1. Capture stream data.
2. Sanitize it to a commit-safe fixture.
3. Replay it with correctness tests.

Keep raw production captures out of commits; only sanitized fixtures should be
checked in.
