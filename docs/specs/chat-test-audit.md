# Chat Test Audit

Scope: `internal/core/chat`, `internal/api/chatclient`, and `internal/app/chat`

Date: 2026-02-26

This audit exists to keep chat testing focused on semantic risk, not just line
coverage. Chat failures are usually lifecycle/scoping failures, so test planning
must prioritize those first.

Use this page as a prioritization guide when adding or rebalancing chat tests.
It should make tradeoffs explicit, not just record an inventory.

## Current strengths

The suite is already strong on stream parsing, turn/round transition handling,
and cancellation cleanup paths. Core rendering contracts (including ANSI/width
constraints) are also covered in high-signal areas.

## Priority gaps

The remaining work is mainly around under-tested edge behavior:

### P0

- `show` tool behavior coverage,
- explicit user-cancel abort path at app boundary,
- mixed lifecycle orchestration scenarios in message list.

### P1

- input bar behavior matrix,
- assistant block update paths,
- query tool truncation metadata edge cases.

### P2

- reducer fuzz/property ordering checks,
- selected golden render snapshots after churn stabilizes.

## Required invariants

No chat test plan is complete unless it protects these invariants:

1. monotonic stream sequencing,
2. strict turn scoping,
3. single terminal state,
4. correct user-cancel persistence semantics,
5. no duplicate or missing tool-result firing.

When choosing between adding breadth and deepening one invariant, prefer
deepening invariant coverage first. These are the failures users feel most.
