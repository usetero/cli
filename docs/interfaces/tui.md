# TUI Interface

The TUI is the long-running interactive runtime for this repository.
When people say “the app,” they usually mean the root Bubble Tea model in
`internal/app`.

The most useful way to think about this interface is as a shell with pages.
The shell owns composition and global routing. Pages own feature behavior.

## What the user-facing contract really is

From the user’s perspective, the TUI guarantees four things:

1. interactions stay responsive (no event-loop blocking),
2. keybindings remain context-correct (page + overlay aware),
3. state transitions are coherent (onboarding -> chat handoff is deterministic),
4. visible data is consistent with local projection/runtime scope.

Everything in the implementation should preserve those guarantees.
If a change violates one of them, users notice immediately, usually as lag,
focus confusion, or seemingly random state jumps.

## How control is split between root and children

The root model (`internal/app/app.go`) is responsible for:

- app-level lifecycle (`quit`, theme changes, overlay open/close),
- routing high-level interaction messages,
- page switching (`onboarding` vs `chat`),
- runtime setup/teardown and layout propagation.

Child models (`onboarding`, `chat`, `statusbar`, `toast`, etc.) are responsible
for their own local update/view logic. They do not reach up into parent state.

This separation is deliberate: it keeps features independent and keeps global
behavior legible in one place.
When you are unsure where new behavior belongs, default to child ownership and
promote to the root only if multiple pages/overlays need to coordinate.

## Sizing and rendering behavior

Layout is not ad hoc in views. The parent computes geometry and pushes it down.

- Fixed-height components get width and report `Height()`.
- Flexible page components get explicit `SetSize(width, height)`.

Rendering is composed in a deterministic order in `app_layout_view.go` and
`app_view.go`, with overlays layered by precedence (drawer, palette, quit).
That predictable layering is part of the interface contract: users should never
see ambiguous focus or inconsistent overlay behavior.

## Message ownership expectations

Shared contracts live in `internal/app/msgs` (for example sync state changes or
drawer prompt forwarding). Feature-local messages remain local unless there is
a real cross-feature ownership need.

Treat messages as facts, not imperative commands. That keeps update routing
composable and makes regression tests meaningful.
In practice, this also makes debugging faster because logs describe what
happened, not what some caller wanted to force.

## Performance and safety expectations

This interface assumes strict Bubble Tea discipline:

- no blocking network/DB work in `View` or synchronous hot `Update` paths,
- side effects only in `tea.Cmd`,
- async completions re-enter update via typed messages.

The root model emits slow-loop telemetry (`slow app update`, `slow app render`)
to catch regressions early. If a change violates these assumptions, users feel
it immediately as input lag.
