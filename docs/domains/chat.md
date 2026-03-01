# Chat Domain

Chat is where this repository is most sensitive to lifecycle mistakes.
The user experience looks simple, but under the hood it combines stream
protocol state, tool execution, persistence, cancellation, and rendering.

If those concerns collapse into one layer, bugs become hard to reason about.
So the code is split intentionally.

## The two-layer model

`internal/api/chatclient` is the protocol/transport adapter layer.
It owns stream wire contracts, parsing, and remote request/response handling.

`internal/core/chat` is the pure chat lifecycle core (for example session/history
state handling and deterministic message-history mutations).

`internal/app/chat` is the UI/runtime shell.
It owns Bubble Tea interaction, focus/layout, round lifecycle wiring, and
persistence integration with the local DB.

That split is the key domain boundary. If you keep it, chat stays testable.

## Dependency rule

Keep dependency direction explicit:

- `internal/app/chat/*` depends on `internal/app/chat/usecase` contracts.
- `internal/app/chat/usecase` owns adapters to `internal/api/chatclient`.
- `internal/core/chat` stays pure and does not import app/api packages.

In practice: if you need `chatclient` in UI files, that is usually a boundary
leak. Add/extend a use-case contract instead.

## What this domain is trying to protect

The expensive failures in chat are usually not visual; they are semantic:

- events from one turn mutating another turn,
- terminal state being applied twice,
- cancellation persisting incorrect assistant output,
- tool-result lifecycle firing too early or too late.

These are domain invariants, not “nice to have” details.

## How to change chat safely

When a change touches transport protocol behavior, start in
`internal/api/chatclient`. When it touches pure lifecycle/state behavior, start
in `internal/core/chat`. Then wire the result into `internal/app/chat` as
message-driven orchestration.

When a change is purely presentation-level, keep it in `internal/app/chat`
without leaking policy into protocol layers.

## Testing posture for this domain

Chat tests should prioritize lifecycle and scoping correctness over incidental
formatting details. See:

- [../operations/testing.md](../operations/testing.md)
- [../specs/chat-test-audit.md](../specs/chat-test-audit.md)

If a bug was user-visible, this domain expects a direct regression test for the
specific semantic failure.
