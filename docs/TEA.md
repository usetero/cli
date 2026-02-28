# TEA Patterns

How to structure Bubble Tea components in this repo.

## Core Model

1. Parent owns layout.
2. Parent propagates size to children.
3. Children render to their assigned region without external positioning logic.

## Sizing Contracts

Two component types:

1. Fixed-height components: `SetWidth(w)` + `Height() int`.
2. Flexible components: `SetSize(w, h)` and render exactly that height.

Rules:

1. Compute layout in `SetSize`, not in `View`.
2. Use `Height` + `MaxHeight` when exact bounds are required.
3. Parent owns spacing between siblings.

## Update Architecture

Prefer split update handlers by message family:

1. Router: top-level `Update` dispatch only.
2. Key handlers: keyboard/focus/scroll interactions.
3. Mouse handlers: selection/click/drag actions.
4. Lifecycle handlers: stream/round/turn transitions.

Use reducers for policy:

1. Reducers are pure (no side effects).
2. Handlers execute side effects (`tea.Cmd`, clipboard, DB via services/models).
3. Reducer tests should cover edge-case transitions.

## Message Design

Message rules for clean TEA boundaries:

1. Every message has one clear owner (the component that defines its semantics).
2. Parent broadcast is allowed; child handling must be explicit and narrow.
3. Timer/poll messages must be namespaced (for example `Source`/`ID`) so one tick cannot trigger unrelated children.
4. `Update` should quickly ignore non-owned messages; do not schedule cmds for foreign messages.
5. Messages should describe facts/events, not commands.

Placement and naming:

1. App-wide contracts: `internal/app/msgs` with `FeatureEvent` naming (example: `SyncStateChanged`).
2. Feature-local contracts: `internal/app/<feature>/msgs`.
3. Private internal messages can stay unexported in the feature package when not shared.
4. Prefer specific names (`ServicesPollTick`) over ambiguous names (`PollMsg`) unless the type includes source identity.

## Projection Rule

If a component has both render math and scroll/hit-test math, compute projection once and reuse it.

Example: message list should use one projection model for:

1. viewport item heights,
2. gaps/dividers,
3. rendered separators.

This avoids render/scroll drift bugs.

## Chat Message List Structure

Current ideal shape in `internal/app/chat/messagelist`:

1. `update.go`: router.
2. `update_key.go`: key side effects.
3. `update_mouse.go`: mouse side effects.
4. `update_lifecycle.go`: lifecycle side effects.
5. `*_reducer.go`: pure policy decisions.
6. `projection.go`: pure layout projection.

## Rendering Guidelines

1. Keep `View` deterministic from current state.
2. Keep rendering tests focused on contracts (width, truncation, markers), not incidental whitespace.
3. Avoid embedding business transitions in rendering methods.

## Test Strategy for TEA Components

1. State transitions: reducer and update tests.
2. Behavior scenarios: message-driven tests over realistic model setup.
3. Rendering contracts: targeted `View` tests.

See [TESTING.md](TESTING.md) for repository-wide standards.
