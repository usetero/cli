# Status Bar

The status bar is a persistent runtime surface, not a decorative header. It has
two jobs:

1. give fast ambient health signals in compact mode, and
2. provide focused drill-down views in the drawer without blocking the main TUI
   loop.

If this surface regresses, users immediately feel it as lag, jitter, or
confusing stale state.

## What the status bar owns

The status bar owns presentation state and interaction for product-surface tabs
that mirror the webapp navigation:

- Control Plane: policies, issues, checks.
- Data Plane: services, log events, edge instances.

It does not own business truth. Facts still come from synced SQLite state and
sync runtime signals.

In code, the root model lives in `internal/app/statusbar/` and each tab package
owns its rendering + tab-local interaction behavior.

## Mental model

Think of the status bar as a shell plus tab plugins.

- Shell (`statusbar.go`, `statusbar_view.go`, `statusbar_drawer.go`):
  compact row layout, drawer open/close, active tab routing, shared key handling.
- Tab contracts (`tabs.go`):
  strict interfaces for tab lifecycle, view rendering, and optional detail
  interaction.
- Shared poll lifecycle (`tabpoll/`):
  typed `PollMsg` -> async fetch -> typed `DataMsg` cycle.
- Product surfaces (`surfaces/`):
  non-interactive summary tabs (policies, issues, checks, log events, edge
  instances) that poll a snapshot, gate on "has data", and render compact +
  drawer views.
- Shared list/detail mechanics (`listdetail/`):
  keyboard navigation and detail-view enter/exit semantics.

This keeps orchestration in one place while pushing tab-specific complexity down
into tab packages.

## Non-negotiable invariants

1. `Update` paths must stay non-blocking. Database work runs in `tea.Cmd`.
2. Status bar view state is presentation-only; it must not become a second
   business domain.
3. Tab-local cache keys are presentation types (primitive fields), not
   `internal/domain` ownership.
4. Drawer interactions are tab-owned; the shell routes keys but does not
   micromanage tab internals.
5. Shared polling/list-detail behavior belongs in `tabpoll` and `listdetail`,
   not duplicated per tab.

## Data and ownership boundaries

Status bar tabs read from local runtime state:

- policy, issue, check, service, and log-event counts from SQLite query
  surfaces,
- sync health from the syncer integration model.

They should not call remote APIs directly from tab update/render paths.
Onboarding handles API-first bootstrap; status bar is runtime projection UI.

## Why the current split exists

Some product surfaces look similar because they solve the same shape of problem:
poll summary state, render compact signals where useful, then offer a drawer
summary for inspection.

The shared `surfaces.Model` exists to remove boilerplate that was previously
easy to drift:

- poll lifecycle bookkeeping,
- "has data" gating,
- snapshot change detection,
- standard compact/drawer rendering.

The remaining logic in each surface should be domain-specific query selection
and snapshot shaping only (the per-surface `fetch...` functions).

## Naming contract

To keep tabs consistent, use these names:

- async message payloads: `...Msg` (for example `detailMsg`),
- cache key builders: `buildStateKey(...)`,
- rendering helpers: `render...` for tab-local view composition.

When two tabs need the same lifecycle behavior, move it to `tabpoll`,
`surfaces`, or `viewkit` instead of introducing one-off names in each tab.

## Practical change checklist

When changing status bar behavior:

1. ensure no synchronous DB/API work was added to `Update` or `View`,
2. keep any new cache/change keys as presentation structs with primitive fields,
3. prefer extending shared helpers before copying lifecycle logic into a tab,
4. run `go test ./internal/app/statusbar/...` and `task do`.
