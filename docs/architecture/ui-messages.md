# UI Messages

This page defines the message contract for the Bubble Tea runtime in the
rebuilt app.

These are architectural rules, not naming trivia. If message ownership becomes
ambiguous, update routing and async behavior become brittle very quickly.

## The message classes

There are four useful kinds of messages in this repo:

1. Bubble Tea input and system messages
   Examples: key presses, window size changes, ticks.
2. child intent messages
   Semantic facts emitted by a child model, such as `Submitted`, `Selected`, or
   `Created`.
3. async result messages
   Results that come back from commands after runtime or external work.
4. shell-level messages
   Root-owned messages such as quit or top-level lifecycle signals.

Every message should have one clear owner.

## Ownership rule

The package that emits a message should usually own that message type.

Keep messages local unless multiple top-level components truly need the same
contract.

That rule keeps message meaning obvious and prevents a central grab-bag package
of vague event types.

## Intent over mechanics

Child models should emit semantic intent messages, not low-level widget facts.

Good:

- `SubmittedMsg`
- `SelectedMsg`
- `CreatedMsg`
- `RefreshRequestedMsg`

Bad:

- `EnterPressedMsg`
- `CursorMovedMsg`
- `TextChangedMsg` when the parent only needs final intent

The parent should not need to know how the child achieved the intent.

## Result messages carry data, not behavior

Async result messages should bring data or errors back into the event loop.
They should not mutate parent state inside closures or hide side effects.

The correct pattern is:

- command does work,
- returns a typed result message,
- parent handles it in `Update`.

This keeps the event loop inspectable and testable.

## Event-loop safety

`Update` and `View` are state/reduction paths.

They must not do:

- network I/O,
- direct database work,
- filesystem I/O,
- sleeps,
- command execution,
- other blocking work.

That work belongs in `tea.Cmd` and must re-enter through typed messages.

## Naming guidance

Use names that describe intent or result shape clearly.

Good local message suffixes:

- `SubmittedMsg`
- `SelectedMsg`
- `CreatedMsg`
- `ResolvedMsg`
- `LoadedMsg`
- `TickMsg`
- `RequestedMsg`

Avoid generic names such as `resultMsg` when multiple async workflows exist in
the same package.

## Routing rule

Parent models should only intercept:

- global keys they explicitly own,
- layout messages they explicitly own,
- typed child intent messages they explicitly own,
- async result messages for work they started.

Everything else should be forwarded to the active child.

That is the routing discipline that keeps parent models from turning into large
switch statements full of accidental coupling.

## Failure patterns

If message design is drifting, you usually see one of these:

- multiple packages handling the same message type,
- parents inspecting low-level child widget behavior,
- commands mutating shared state outside `Update`,
- generic result messages that hide which workflow completed,
- children depending on sibling-specific messages.

Those are architecture bugs, not just naming issues.

## Fast code entry points

- [`internal/interfaces/tui/root/model.go`](../../internal/interfaces/tui/root/model.go)
- [`internal/interfaces/tui/screens/onboarding/model.go`](../../internal/interfaces/tui/screens/onboarding/model.go)
