# System Overview

The CLI is the presentation runtime for Tero, not the product authority.
That distinction is the most important mental model in this repository.

The control plane owns business rules and authoritative state transitions. This
codebase owns terminal UX, local runtime orchestration, and responsive rendering
on top of a synced local projection.

## What Runs Here

One executable (`cmd/tero/main.go`) exposes three user-facing interfaces:

- the interactive TUI (`internal/app`)
- direct commands (`internal/cmd`)
- an MCP adapter surface (planned, intentionally thin)

All dependency composition happens at startup in `internal/cmd/root.go`. Feature
packages should not self-compose service graphs.

## How a Session Actually Flows

A normal runtime starts in onboarding, not chat. The app gathers bootstrap facts
through deterministic gates, then initializes the account-scoped runtime, then
hands off to chat.

In code, that lifecycle starts at `cmd/tero/main.go`, is wired in
`internal/cmd/root.go`, enters the app state machine in `internal/app/app.go`,
drives onboarding orchestration, opens the runtime database, starts sync and
uploader, and only then settles in the steady-state chat surface.

The important architectural point is the handoff boundary: onboarding is a
bootstrap workflow; chat is a synced-runtime workflow. Do not blur them.

## State Ownership

Authoritative truth remains remote. Local SQLite is a projection for fast local
reads and immediate UI responsiveness. PowerSync and uploader converge local and
remote state over time.

Onboarding is the intentional exception: until account/runtime prerequisites are
ready, it uses API-driven bootstrap. After completion, the UI should read from
the local projection and write through the local mutation pipeline.

## Package Boundaries

The boundaries are deliberately simple:

- `cmd/` and `internal/cmd/` compose dependencies and start surfaces.
- `internal/app/` and feature subpackages render and coordinate UI state.
- `internal/core/bootstrap/` owns deterministic onboarding transition policy.
- `internal/core/chat/` owns pure chat lifecycle/session policy.
- service/domain packages (`internal/boundary/graphql`, `internal/boundary/chat`,
  `internal/auth`, `internal/domain`, `internal/preferences`) expose contracts
  and adapters.
- data/sync packages (`internal/sqlite`, `internal/powersync`, `internal/upload`)
  own projection storage and convergence mechanics.

When in doubt, move policy inward and keep presentation code thin.

## Invariants That Must Stay True

These are hard constraints, not style preferences:

- composition happens at the top (`cmd/`), not inside feature packages
- presentation depends inward; core/service packages never depend on UI
- cross-feature UI communication uses explicit typed messages
- no blocking network/database work in `View` or hot synchronous `Update` paths
- scoped clients are derived immutably (`WithAccountID(...)`), not mutated shared
  globals

## Environment and Tenant Safety

Runtime state is environment-scoped and org-scoped. Config comes from
`internal/config`, active organization comes from preferences, and local storage
paths are isolated per org/environment. This is what prevents cross-tenant bleed
in long-running terminal sessions.

## Where To Change Behavior

Change this repository when the work is about UX flow, rendering, input behavior,
runtime orchestration, or local developer ergonomics.

Change the control plane when the work is about product policy, business rules,
authoritative state transitions, or semantics that should be true across all
clients.

## Common Failure Modes

Most regressions come from the same mistakes: putting business policy in UI
handlers, treating SQLite as authority instead of projection, mutating shared
scoped clients, doing blocking I/O in the event loop, or letting messages cross
feature boundaries without clear ownership.
