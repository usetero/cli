# Status Bar

The status bar is the shell-level orientation surface for the TUI.

In the rebuilt app, it is deliberately much smaller in scope than the legacy
tabbed status subsystem. It does not own drawers, detail views, or product
queries. It presents brand, environment, and session sync state in one compact
line.

## Why this surface matters

The status bar is always visible. If it is wrong, users immediately lose trust
in the rest of the app:

- stale "ready" signals hide sync failures,
- noisy state labels make the shell feel unstable,
- width regressions break the visual rhythm of the app header,
- presentation logic leaking upward makes the root shell brittle.

So even though the component is small, the contract matters.

## The current model

The status bar in
[`internal/interfaces/tui/components/statusbar`](../../internal/interfaces/tui/components/statusbar)
has one job: render session status from [`internal/runtime/session`](../../internal/runtime/session)
into a width-aware shell header.

It reads:

- environment name,
- runtime `Running` state,
- PowerSync sync state exposed through session status.

It does not fetch, poll, or derive product truth on its own.

## What must stay true

- runtime session state is the source of truth for header sync status,
- the status bar remains presentation-only,
- compact rendering degrades intentionally as width shrinks,
- production environment stays visually quieter than non-production shells,
- root shell owns placement; the status bar owns only its rendered line.

## Failure behavior

Most problems here fall into one of three buckets:

- wrong status label: inspect session status mapping before touching theme code,
- visual overflow or truncation: inspect width-aware rendering in the component,
- stale shell state: inspect runtime/session publication, not the header first.

If the status bar ever needs polling, data loading, or complex local state, that
is almost certainly a boundary regression.

## Why the legacy mental model no longer applies

The old status bar page described a much larger feature with tabs, drawers, and
query lifecycles. That is not the rebuilt app.

Today the root shell owns a single header slot, and the status bar is a focused
presentational component inside that slot. Treating it like a mini-application
would overcomplicate both the code and the docs.

## Code entry points

- [`internal/interfaces/tui/components/statusbar/model.go`](../../internal/interfaces/tui/components/statusbar/model.go)
- [`internal/interfaces/tui/components/statusbar/sync_presenter.go`](../../internal/interfaces/tui/components/statusbar/sync_presenter.go)
- [`internal/interfaces/tui/root/model.go`](../../internal/interfaces/tui/root/model.go)
- [`internal/runtime/session/service.go`](../../internal/runtime/session/service.go)
