# Observability

Observability in this repo is intentionally simple: structured logs, explicit
user-facing notifications, and invariant-driven tests.

The goal is not “more telemetry.” The goal is fast, reliable diagnosis when
interactive behavior regresses.

## What signals matter most

In practice, you will spend most of your time with:

- scoped application logs,
- onboarding gate telemetry,
- syncer/uploader lifecycle logs,
- slow-loop update/render logs,
- user-facing toast output for surfaced failures.

These signals are enough to explain most failures without adding tracing
infrastructure complexity.

## Triage workflow that works

Start from one concrete user symptom and keep the loop tight:

1. Reproduce in `dev` (or `local`) with minimal noise.
2. Tail the relevant environment logs.
3. Map observed transitions to expected architecture invariants.
4. Isolate root cause to one boundary.
5. Add regression coverage at that boundary.

The failure mode to avoid is collecting broad logs without a hypothesis.

## Environment-aware log reading

Use the right environment path for your reproduction:

```bash
tail -f ~/.tero/environments/dev/tero.log
```

Switch `dev` to `local`/`prd` intentionally.
Do not mix environments during diagnosis unless you are explicitly comparing behavior.

## What “good observability” looks like here

A behavior issue should be diagnosable by answering what state the app believed
it was in, what message/event triggered the transition, whether local runtime
dependencies were ready, and whether user-facing feedback matched internal
state.

If those answers are not visible in logs/tests, we should improve instrumentation
at the relevant boundary, not scatter generic logging everywhere.
