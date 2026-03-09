# TUI Interface

The TUI is the interactive presentation surface for the rebuilt app. It should
feel responsive and cohesive, but it is still only a surface over domain,
infrastructure, and runtime layers.

If you are changing TUI behavior, start with these architecture pages:

- [`../architecture/ui-runtime.md`](../architecture/ui-runtime.md)
- [`../architecture/ui-messages.md`](../architecture/ui-messages.md)
- [`../architecture/ui-layout.md`](../architecture/ui-layout.md)
- [`../architecture/theme-and-chrome.md`](../architecture/theme-and-chrome.md)

This page is narrower. It describes the TUI-specific contract as a user-facing
surface.

## What the TUI owns

The TUI owns:

- interactive terminal behavior,
- page and flow composition,
- local widget interaction,
- routing between screens,
- rendering and shell chrome,
- translating user intent into runtime or domain calls.

The TUI does not own product truth, long-running business workflows, or
concrete transport/storage implementation.

## Current structure

The current surface is composed from:

- [`internal/interfaces/tui/app.go`](../../internal/interfaces/tui/app.go):
  TUI startup and composition
- [`internal/interfaces/tui/root`](../../internal/interfaces/tui/root):
  root shell
- [`internal/interfaces/tui/present`](../../internal/interfaces/tui/present):
  typed content and surface rendering
- [`internal/interfaces/tui/screens/onboarding`](../../internal/interfaces/tui/screens/onboarding):
  onboarding flow
- [`internal/interfaces/tui/components`](../../internal/interfaces/tui/components):
  reusable interactive widgets
- [`internal/interfaces/tui/chrome`](../../internal/interfaces/tui/chrome):
  shell layout and brand helpers
- [`internal/interfaces/tui/theme`](../../internal/interfaces/tui/theme):
  semantic presentation tokens

## Surface rules

The TUI should follow these rules consistently:

- root owns shell concerns
- flow models own routing and runtime calls
- leaf models own local interaction state only
- all async work re-enters through typed messages
- chrome owns shell and frame layout
- present owns shared content surfaces
- components own reusable interaction behavior

If a change breaks one of those rules, it is usually being made at the wrong
layer.

## Input and terminal policy

The root view owns global terminal policy:

- `AltScreen` enabled
- `WindowTitle` set
- `MouseMode` disabled by default unless a surface truly needs it

Program startup applies the shared input filter under
[`internal/interfaces/tui/filter`](../../internal/interfaces/tui/filter) to
reduce noisy terminal input bursts.

## Layout contract

The shell has three conceptual regions:

- header
- body
- footer

Header and footer are shell-owned. Body placement is handled by chrome and page
frame rules, not by random per-screen padding.

That layout contract is described in
[`../architecture/ui-layout.md`](../architecture/ui-layout.md).

## Testing expectation

TUI tests should protect:

- route correctness,
- message ownership,
- child/parent forwarding behavior,
- user-visible rendering contracts when layout or visibility matters.

For broader confidence, use the executable smoke and live integration lanes in
`cmd/tero`.

## Failure patterns

The TUI surface is drifting when:

- screens start performing their own service or runtime work,
- local widget state is treated as product truth,
- shell layout is patched per screen instead of through chrome,
- user-visible behavior depends on hidden global state outside the event loop.

Those failures usually mean the wrong layer owns the change.
