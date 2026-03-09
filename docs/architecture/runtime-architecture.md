# Runtime Architecture

This page defines the layer split that sits underneath all user-facing
interfaces.

The goal is simple: when a new engineer opens this repository, they should be
able to decide where a behavior belongs before they touch code.

## The three non-UI layers

Before the TUI or command surfaces exist, the system is already mostly defined
by three layers:

1. `internal/domains`
2. `internal/infrastructure`
3. `internal/runtime`

If those layers are clean, the UI becomes straightforward. If those layers are
muddy, the UI turns into whack-a-mole.

## Domain responsibilities

Domain packages own:

- typed business inputs and outputs,
- validation and normalization,
- service contracts,
- mapping between user intent and business-shaped operations.

Domain code should answer questions like:

- what inputs are valid,
- what operations exist,
- what state transitions are legal at a business level.

Domain code should not answer:

- how HTTP is performed,
- where data is stored,
- how the TUI is arranged on screen.

## Infrastructure responsibilities

Infrastructure packages own:

- concrete clients and storage adapters,
- HTTP, SQLite, PowerSync, WorkOS, keyring, and preferences implementations,
- low-level translation between external protocols and domain contracts.

The right mental model is "focused libraries." These packages should be narrow,
concrete, and boring.

Infrastructure code should not:

- become product policy,
- coordinate long-running workflows,
- know about screen routing or UI presentation.

## Runtime responsibilities

Runtime packages own:

- long-running state machines,
- orchestration across multiple services,
- lifecycle management,
- polling, readiness, session transitions, and bootstrap progression.

Runtime is where "what should happen over time?" lives.

Examples in this repo:

- onboarding progression and projection in
  [`internal/runtime/onboarding`](../../internal/runtime/onboarding)
- account-scoped lifecycle in
  [`internal/runtime/session`](../../internal/runtime/session)

Runtime packages sit above domain services and below user-facing interfaces.

## How failures should be interpreted by layer

Each layer should fail in a different way:

- domains reject invalid business input early and clearly,
- infrastructure returns concrete transport/storage/protocol failures,
- runtime translates repeated low-level failures into lifecycle state,
- interfaces render or route based on that state.

If a UI screen is parsing HTTP details, or infrastructure is deciding product
recovery policy, ownership has drifted.

## The key ownership rule

Use this question to place code:

"Is this about validity, implementation, or coordination over time?"

- validity -> domain
- implementation detail -> infrastructure
- coordination over time -> runtime

That one question resolves most placement arguments.

## Constructor and validation policy

This repo now follows a strict split:

- constructor invariants live in constructors,
- request and mutation validation live in typed inputs and `Validate`,
- long-running behavior checks live in runtime state machines.

This keeps responsibilities separated:

- domains validate business inputs,
- infrastructure validates concrete adapter inputs,
- runtimes validate lifecycle and state progression.

## What must stay true

- domains expose typed operations, not UI-shaped helpers,
- infrastructure packages remain single-purpose and concrete,
- runtime owns coordination over time and retry/readiness policy,
- constructors own wiring invariants,
- request structs own mutation validation.

## What not to leak upward

User-facing surfaces should not have to know:

- how tokens are stored,
- how GraphQL requests are made,
- how PowerSync or uploader loops are wired,
- how preferences are persisted,
- how onboarding state is recomputed.

If interface code needs to know those details, the lower layers are not clean
enough yet.

## What not to push downward

Lower layers should not know:

- what key was pressed,
- which screen is active,
- how chrome is arranged,
- how error cards are rendered,
- how a help bar is styled.

That belongs to interface code.

## Code paths to read

Start here when learning the rebuilt app:

- [`internal/domains`](../../internal/domains)
- [`internal/infrastructure`](../../internal/infrastructure)
- [`internal/runtime`](../../internal/runtime)
- [`internal/interfaces/tui/compose_onboarding.go`](../../internal/interfaces/tui/compose_onboarding.go)

The last file is a useful bridge because it shows how the TUI composes the
lower layers without owning their logic.
