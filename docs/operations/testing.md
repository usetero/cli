# Testing

Testing in this repo is about protecting behavioral invariants, not maximizing
line coverage. Coverage can be a useful smell detector, but the contract we
care about is “does runtime behavior stay correct under real message flow?”

## What to optimize for

Prefer tests that lock down cross-component behavior:

- deterministic state transitions,
- message ownership and scoping,
- concurrency safety under runtime switching,
- user-visible failure/recovery semantics.

If a change can break a user flow but has no regression test, the test suite is
incomplete even if coverage looks good.
The test suite should answer "will this feel correct to users?" not just
"did lines execute?"

## Practical test strategy

Use a layered approach:

1. pure reducers/state transitions first,
2. component update/message-flow tests second,
3. rendering contract tests where layout/visibility behavior matters,
4. integration/correctness suites for protocol-level guarantees.

Mock only where boundaries are truly external or expensive.
For local dependencies like SQLite/log/theme, real instances are often clearer
and more reliable than deep mocks.
The bias here is toward realistic behavior and lower false confidence.

## High-risk areas that must stay protected

Chat and onboarding both rely on strict lifecycle semantics.
The most important invariants include:

- turn scoping and stream ordering,
- single terminal outcome semantics,
- cancellation behavior and persistence rules,
- deterministic onboarding transitions and safe gate navigation.

When a bug is fixed in these areas, add a focused regression test in the same
change.
That practice has the highest long-term ROI for stability.

## Command workflows

For normal local validation:

```bash
task do
```

For architecture/lifecycle guardrails only:

```bash
task lint:naming
task lint:architecture
```

For chat-focused work:

```bash
go test ./internal/core/chat ./internal/boundary/chat ./internal/app/chat/... -count=1
```

For explicit integration or correctness runs:

```bash
task test:integration
task test:correctness
task test:correctness:powersync-replay
```

Use these targeted workflows when debugging protocol/sync behavior that unit
tests alone cannot validate.
