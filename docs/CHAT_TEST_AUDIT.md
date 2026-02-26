# Chat Test Audit

Scope: `internal/chat` and `internal/app/chat`

Date: 2026-02-26

## Current Snapshot

Coverage (package-level):

- `internal/chat`: 72.7%
- `internal/app/chat`: 37.8%
- `internal/app/chat/messagelist`: 32.1%
- `internal/app/chat/messagelist/round`: 60.9%
- `internal/app/chat/messagelist/round/turn`: 50.3%
- `internal/app/chat/messagelist/round/turn/assistant`: 18.9%
- `.../assistant/blocks/tools/action`: 83.3%
- `.../assistant/blocks/tools/query`: 35.1%
- `.../assistant/blocks/tools/show`: 0.0%

Coverage alone is not the goal, but these numbers highlight where behavior risk is concentrated.

## What Is Strong

1. Core stream parsing and accumulation in `internal/chat` has solid unit coverage.
2. Turn/round state transitions (including tool-loop dedupe) are covered.
3. Cancel cleanup and DB orphan handling are covered in app-level tests.
4. Width and ANSI render contracts are covered for key components.

## Gaps To Close

Priority P0:

1. `show` tool behavior has no tests (`0.0%`).
2. App-level test for user-cancel path with explicit abort intent (`user_cancelled`) through snapshot API.
3. Broader tests for `messagelist` orchestration under mixed mouse/keyboard and stream lifecycle.

Priority P1:

1. Input bar behavior coverage is missing (`0.0%`): submit/newline/palette/restore pending text.
2. Assistant block creation/update paths are relatively thin (`18.9%`).
3. Query tool edge rendering (large row counts + truncation metadata) is partially covered.

Priority P2:

1. Add fuzz/property tests for reducer event ordering constraints.
2. Add golden-style snapshots for selected complex render outputs if churn stabilizes.

## Required Invariant Suite

These must stay tested as first-class invariants:

1. Stream event sequencing (`seq` monotonic per turn).
2. Turn scoping for stream and tool events.
3. Terminal state uniqueness (`completed` | `tool_use` | `aborted` | `failed`).
4. User-cancel is non-error and non-persisted as committed assistant turn.
5. Non-user abort can persist assistant partial with `stop_reason=aborted`.
6. No duplicate tool result firing.
7. Conversation history stays valid after cancel/retry/failure.

## Execution Plan

1. Add `show` tool tests first (P0).
2. Add explicit user-cancel abort app-level test (P0).
3. Add input bar test suite (P1).
4. Add reducer fuzz/property checks (P2).

Run gate:

```bash
go test ./internal/chat ./internal/app/chat/... -count=1
```

