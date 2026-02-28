# UI Architecture

The easiest way to understand the TUI is to start at the root model in
`internal/app/app.go` and follow one update cycle all the way through.

The root model is intentionally conservative. It owns composition, routing, and
runtime orchestration, while child models own feature behavior. That split is
what keeps the app debuggable as features evolve.

## How one update cycle works

`Model.Update` is structured as a router, not a business-logic blob.
It handles messages in a strict order: global lifecycle first (`quit`, palette
open/close, theme set), then direct interaction routing (window resize,
key/mouse), then onboarding orchestration messages (runtime ensure, completion
handoff), and finally child update fanout (statusbar, toast, current page).

That order matters. It ensures overlays and global actions are resolved before
feature updates, and it keeps the current page model isolated from app-wide
control flow.

## The root owns shell concerns, not feature internals

The root app model owns shell concerns: active page (`onboarding` vs `chat`),
overlays (drawer, palette, quit dialog), layout composition, and runtime
session setup/teardown.

Feature models own feature concerns: their state machines, local input
semantics, and feature-specific rendering.

This is why `updateChildren` simply forwards the same message to relevant
children rather than trying to interpret child state in the parent.

## Layout is computed, not improvised in views

Layout behavior is centralized in `app_layout_view.go`.
The app computes available content size once, then pushes explicit sizes into
children via `SetWidth`/`SetSize`.

The view path then composes a render frame in a fixed order:

toast -> status bar -> page -> key bar

Overlays are layered after base rendering (`drawer`, then `palette`, then
`quit dialog` precedence). This keeps overlay behavior predictable and avoids
feature models trying to coordinate z-order with each other.

## Message contracts are explicit by scope

Shared app-level messages live in `internal/app/msgs` (for example
`SyncStateChanged`, `DrawerPrompt`). Feature-local messages stay in their own
feature packages.

In practice, this gives you a clean rule of thumb:

if multiple top-level components need to understand a message, define it in
`internal/app/msgs`; otherwise keep it local.

This prevents accidental coupling and keeps message ownership obvious.

## Side effects stay in commands, not render/update hot paths

The architecture assumes Bubble Tea’s separation of state and effects:

`View` should be deterministic from model state. Expensive or external work
should run in `tea.Cmd`, and async completions should return typed messages
that re-enter `Update`.

You can see this pattern in onboarding steps, status polling, sync state
notifications, and runtime initialization handoff.

## Onboarding inside the UI runtime

Onboarding is treated as a page model with its own orchestrator.
The app does not micromanage onboarding internals; it only reacts to explicit
bootstrap contracts (for example `EnsureRuntime`, `OnboardingComplete`).

That keeps responsibilities clean:

onboarding decides when bootstrap facts are complete, and the app decides when
to initialize runtime and switch to chat.

## Performance posture

The root model logs slow update/render loops (`perf_logging.go`) with enough
context to diagnose blocked event-loop behavior. This is important in terminal
UIs because a small synchronous regression can feel like a freeze.

The architecture expectation is simple: never do blocking network/DB work in
`Update` or `View`. If you follow that rule and keep message ownership narrow,
the UI remains responsive under real workloads.
