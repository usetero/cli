# Observability

Use this doc for diagnosing runtime behavior.
For coding conventions around log emission, see [LOGGING.md](LOGGING.md).

## Signals

Primary signals in this repo:

1. Structured application logs.
2. User-facing toasts for surfaced failures.
3. Test failures in targeted subsystems.

## Fast Triage Workflow

1. Reproduce with minimal scope.
2. Follow scoped logs from root component to failing child.
3. Correlate lifecycle transitions with expected invariants.
4. Confirm user-visible impact (toast/state/output).
5. Add or update regression tests once root cause is confirmed.

## Local Debugging

Use the local environment log stream during reproduction:

```bash
tail -f ~/.tero/environments/prd/tero.log
```

Then narrow by scope and identifiers (for example `scope=app/chat`, `turn_id`, `conversation_id`).

## Chat Incident Checklist

1. Did stream events arrive out of order for a turn?
2. Did an event for one turn mutate another turn?
3. Was terminal outcome ambiguous or duplicated?
4. Was user-cancel treated as error/persisted incorrectly?
5. Did stale tool completion mutate active state?

Map findings to tests in `internal/chat` and `internal/app/chat/...`.

## Escalation Guidance

Escalate when:

1. State invariants are violated across turns.
2. Data persistence behavior contradicts policy.
3. User-visible regressions cannot be explained by local environment issues.

When escalating, include:

1. Repro steps.
2. Relevant scoped log excerpts.
3. Expected vs actual behavior.
4. Candidate invariant violated.
