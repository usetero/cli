# Data Flow

This CLI intentionally uses different data flows at different phases of the
session. If you miss that phase split, the code can look inconsistent. Once you
see it, the behavior is straightforward.

## Start with the key rule

Remote control plane state is authoritative. Local SQLite is a projection used
for responsiveness, not a competing source of truth.

Everything in this repo follows from that rule.

## There are three practical phases

When an engineer says “data flow” in this repo, they usually mean one of three
phases: command execution (`internal/cmd`), onboarding bootstrap
(`internal/app/onboarding`), or post-bootstrap runtime
(`internal/app` + SQLite + PowerSync + uploader).

Those phases use different paths on purpose.

## Phase 1: direct command flow

For plain CLI commands, the flow is direct:

parse input -> call API service -> render output.

You can see this in `internal/cmd/*` and API adapters in `internal/api`.
There is no local projection dependency for command correctness.

## Phase 2: onboarding bootstrap flow

Before account runtime is initialized, the app does not have an account-scoped
local DB session running. So onboarding is API-first and deterministic.

Steps emit facts, those facts become canonical bootstrap events, and the
transition engine advances gates. During this phase, the orchestrator keeps
bootstrap state locally and scopes API services as account state appears.

Relevant code:

- `internal/app/onboarding/transition_*.go`
- `internal/core/bootstrap/*`
- `internal/app/onboarding/service_scope.go`

This is why onboarding can run cleanly even before sync/database runtime is
ready.

## Phase 3: runtime data flow after onboarding

Once onboarding reaches runtime initialization:

1. the app opens account-scoped SQLite (`runtime_database.go`),
2. starts PowerSync (`runtime_sync.go`),
3. starts uploader against the local CRUD queue (`runtime_sync.go` + `internal/upload`),
4. derives scoped API/chat clients with immutable account-scoped instances (`WithAccountID`),
5. and reads status/chat data from the local DB.

At that point, the UI behaves like a local-first runtime backed by sync.

## What PowerSync and uploader each do

PowerSync and uploader solve opposite directions of convergence.

PowerSync is mostly remote -> local:

- sync stream drives local projection freshness,
- app components consume sync state (`SyncStateChanged`) and local tables.

Uploader is mostly local -> remote:

- watches local CRUD queue,
- invokes table-specific handlers against APIs,
- completes write checkpoints,
- notifies syncer via `NotifyUploadCompleted` so both sides converge.

This handshake is what prevents local writes from drifting out of protocol with
the sync stream.

## Why account scoping is explicit everywhere

Multi-tenant correctness depends on scope discipline.

The app derives scoped clients rather than mutating shared global state.
You can see this in `APIServices.WithAccountID(...)` and chat
`WithAccountID(...)` usage.

On org/account switches, runtime is torn down, scope is cleared, and onboarding
restarts from the right gate. That avoids cross-account leakage in long-lived
terminal sessions.

## Where engineers usually make mistakes

The repeated failures in this area are predictable:

- treating local SQLite as authority instead of projection,
- adding onboarding behavior that assumes sync runtime exists too early,
- mutating shared clients instead of deriving scoped copies,
- bypassing uploader/checkpoint flow for local mutations.

If you keep phase boundaries explicit, those mistakes disappear.
