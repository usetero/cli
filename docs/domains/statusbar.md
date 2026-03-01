# Status Bar

The status bar is a persistent runtime surface, not a decorative header. It has
two jobs:

1. give fast ambient health signals in compact mode, and
2. provide focused drill-down views in the drawer without blocking the main TUI
   loop.

If this surface regresses, users immediately feel it as lag, jitter, or
confusing stale state.

## What the status bar owns

The status bar owns presentation state and interaction for five tabs:
waste, quality, compliance, services, and sync.

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
- Shared policy-tab behavior (`policytab/`):
  reusable base for waste/quality/compliance polling, change detection, and
  list/detail cursor lifecycle.
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
5. Shared polling/list-detail behavior belongs in `tabpoll`, `policytab`, and
   `listdetail`, not duplicated per tab.

## Data and ownership boundaries

Status bar tabs read from local runtime state:

- policy/service counts from SQLite query surfaces,
- sync health from the syncer integration model.

They should not call remote APIs directly from tab update/render paths.
Onboarding handles API-first bootstrap; status bar is runtime projection UI.

## Why the current split exists

Waste, quality, and compliance look similar because they solve the same shape of
problem: poll summary + categories, render compact signal, then offer a
list/detail drawer for category inspection.

The shared `policytab.Base` exists to remove boilerplate that was previously
easy to drift:

- poll lifecycle bookkeeping,
- "has data" gating,
- cursor clamp and state-change checks,
- standard list/detail navigation wiring.

The remaining logic in each tab should be domain-specific rendering and query
selection only.

## Naming contract

To keep tabs consistent, use these names:

- async message payloads: `...Msg` (for example `detailMsg`),
- cache key builders: `buildStateKey(...)`,
- rendering helpers: `render...` for tab-local view composition.

When two tabs need the same lifecycle behavior, move it to `tabpoll`,
`policytab`, or `viewkit` instead of introducing one-off names in each tab.

## Practical change checklist

When changing status bar behavior:

1. ensure no synchronous DB/API work was added to `Update` or `View`,
2. keep any new cache/change keys as presentation structs with primitive fields,
3. prefer extending shared helpers before copying lifecycle logic into a tab,
4. run `go test ./internal/app/statusbar/...` and `task do`.
