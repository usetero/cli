# Onboarding Bootstrap

Onboarding is the bootstrap runtime that gets a user from “fresh session” to
“chat can safely run on a real account/workspace runtime.”

The important thing to understand is that this is not a linear wizard. It is a
deterministic state machine with a UI wrapped around it.

## Why onboarding is separate from normal app runtime

Before runtime initialization, the app does not yet have a valid account-scoped
local database + sync loop. That means the normal “local SQLite projection”
model is not ready yet. During bootstrap, onboarding uses API-driven steps to
resolve auth, org, account, Datadog setup, workspace, and sync readiness.

Only after bootstrap is complete does the app hand off to chat/runtime state.

You can see that handoff in `internal/app/onboarding_orchestration.go`, where
`bootstrap.OnboardingComplete` switches app state to chat and initializes the
chat model.

## The mental model that makes this code predictable

Think of onboarding as five collaborating parts.

- `bootstrap.State`: accumulated bootstrap truth (user, org, account, workspace,
  Datadog readiness).
- Step messages: typed facts (`RoleSelected`, `WorkspaceSelected`,
  `SyncComplete`) emitted from gate components.
- Event adapter: orchestrator converts step facts into canonical
  `bootstrap.Event` values.
- Transition engine: `bootstrap.ApplyEvent`, pure and deterministic for the same
  `(state, event)` input.
- Orchestrator + steps: `internal/app/onboarding` owns gate selection/navigation;
  each step owns only its gate UI and async side effects.

That split is intentional: policy is centralized, while UI/effects are pushed to step packages.

## How a transition actually flows at runtime

A typical update cycle looks like this:

1. A step emits a typed bootstrap message.
2. The orchestrator converts it with `bootstrap.EventFromMessage`
   (`transition_policy.go`).
3. The orchestrator applies the transition with `bootstrap.ApplyEvent`
   (`transition_apply.go`).
4. The transition result is executed in `transition_cmds.go` (advance, complete,
   or noop).
5. Gate navigation creates and initializes the next step
   (`gate_navigation.go`, `gate_definitions.go`).

Two design details matter here:

- Gate rewinding is explicit. If prerequisites for a requested gate are missing,
  `rewindGateFor` moves back to the earliest valid gate.
- Unsupported gates are handled safely (logged + no-op), not by panic.

## Step contract and visibility behavior

The step contract is intentionally explicit (`model_interfaces.go`): lifecycle,
help bindings, and visibility/status methods are all part of one interface.

That gives the orchestrator one consistent way to render any gate:

- If a step is visible, render the step view.
- If a step is hidden, render the step’s status text.

There is no generic hidden fallback anymore. Hidden steps must provide their own
status, which is why “stuck on getting ready” is now diagnosable and fixable.

## Observability and jitter/stuck diagnosis

Onboarding now emits gate telemetry in `gate_telemetry.go`:

- gate enter: `gate`, `trigger`
- gate exit: `gate`, `next_gate`, `trigger`, `duration_ms`

This gives you a concrete sequence and timing for every gate transition.
When users report jitter or stalls, this telemetry tells you which gate they are
in, which event moved them, how long each gate took, and whether completion was
reached with valid state.

## Non-negotiable invariants

The architecture depends on these rules staying true:

1. Transition policy is deterministic and centralized in `internal/core/bootstrap`.
2. Step components emit facts, not routing commands.
3. `Update`/`View` paths do not perform blocking I/O.
4. Bootstrap state is local to onboarding until completion handoff.
5. Gate navigation is safe under missing prerequisites (rewind) and unsupported gates (log/no-op).

If you keep those intact, onboarding stays predictable even as step UIs evolve.
