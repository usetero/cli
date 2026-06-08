# Plan: Drop PowerSync from the CLI

Status: **proposal / scoping** (no code yet). Target: dedicated branch, separate
from `clay-mirror-new-ui-surfaces`.

## Why

The control plane has moved off PowerSync. Migration
`20260424084500_remove_status_cache_tables.sql` (control-plane, 2026-04-24)
drops `datadog_account_statuses_cache`, `service_statuses_cache`,
`log_event_statuses_cache`, and `issue_statuses_cache` — described in the
migration itself as *"the former PowerSync materialization tables"* — and routes
status reads through canonical Postgres views instead.

The CLI's synced schema (`internal/sqlite/schema.sql`,
`internal/powersync/extension/schema.json`) was last regenerated **2026-02-25**
and is still built entirely around those dropped `*_cache` tables. So the CLI is
syncing a schema the control plane no longer maintains. PowerSync is effectively
dead weight for reads today.

This dovetails with the original task (adopt the latest control-plane APIs):
the new first-class entities — `issues`, `checks`, `edgeInstances` — are exposed
over **GraphQL**, not PowerSync. Moving the CLI's reads to direct GraphQL both
removes the stale sync engine and picks up the new APIs in one motion.

Aligns with `CLAUDE.md`: *"CLI is presentation only,"* local data is
*"queried (CLI)"*. The traditional CLI commands already query GraphQL directly;
this brings the TUI in line.

## What PowerSync actually does in the CLI today

The local `sqlite.DB` **is** a PowerSync-extension-loaded SQLite connection
(`internal/sqlite/database.go` loads `sqlite3_powersync_init`). PowerSync plays
three roles through that connection:

1. **Local store / schema** — the synced tables and views are PowerSync-managed.
2. **Download (read path)** — replicates control plane → local SQLite. The
   status bar, onboarding, and chat history all read these tables.
3. **Upload outbox (write path)** — local writes through PowerSync views are
   captured into the `ps_crud` queue. The `internal/upload` uploader drains
   `ps_crud` and translates each entry into a **GraphQL mutation**
   (`conversationHandler`→`graphql.Conversations`, `messageHandler`→message
   mutations, `policyHandler`→`graphql.Policies`, `serviceHandler`→
   `graphql.Services`).

Key insight: the upload **transport is already GraphQL**. PowerSync is only the
local outbox queue in front of it. The read path, by contrast, depends wholly on
PowerSync replication.

## The one decision that shapes everything: local store or stateless?

Because `sqlite.DB` is the PowerSync extension, removing PowerSync forces a
choice about the local store:

- **Option A — Stateless (direct GraphQL, no local DB).** TUI reads issue
  GraphQL queries on demand; chat renders from in-memory state and persists via
  GraphQL mutations. Delete the local SQLite store entirely. Simplest end state,
  best alignment with "CLI is presentation only," but changes chat's optimistic
  persistence model and gives up offline/local cache.
- **Option B — Plain SQLite cache + custom outbox.** Keep a local SQLite (no
  PowerSync extension) as a read cache and write outbox; populate it from
  GraphQL queries and drain a hand-rolled outbox to GraphQL mutations. Preserves
  offline behavior but re-implements a meaningful slice of PowerSync.

**Decision: Option A (confirmed 2026-06-08).** The control plane is the source
of truth, the CLI is presentation, and the GraphQL client + mutations already
exist. Resolved gating questions:

- **Offline: not required.** No local cache needed; safe to delete the local DB.
- **Chat history: available over GraphQL.** Dropping local SQLite loses no
  history — conversations/messages are served by the control plane.
- **Edge instances: in scope.** Wire `edgeInstances` GraphQL in phase 1
  alongside issues/checks (no longer a deferred stub).

Optimistic chat UI is preserved with in-memory state reconciled on mutation
success/failure. Option B is off the table.

The rest of this plan assumes **Option A**.

## Schema bump findings (2026-06-08, regenerated mirror)

Regenerating `gen/schema.graphql` against the live control plane (it was ~2
months stale) surfaced breaking changes — the current operations no longer
validate:

| Existing CLI op | Status in current control-plane schema |
|---|---|
| `createConversation` / `updateConversation` / `deleteConversation` | **removed** — chat is not a control-plane GraphQL concern |
| `createMessage` | **removed** |
| `approveLogEventPolicy` / `dismissLogEventPolicy` | **removed** → policy lifecycle moved to the Issue model (`ignoreIssue`, `createLogEventPolicy`) |
| `updateService` | renamed → `setServiceEnabled` |
| `workspaces` (query) | **removed** |

Consequence: the **current uploader is already broken** against the live
control plane (pushes `createConversation`/`createMessage` mutations that no
longer exist). Rebuilding the data layer is not optional cleanup.

**Chat is ephemeral (decided 2026-06-08).** Conversation/message history is no
longer persisted across sessions — in-memory during a session is enough. So:
delete all chat persistence (conversation/message GraphQL ops, uploader
handlers, sqlite conversations/messages tables and query surfaces); chat state
lives in-memory and streams via `internal/boundary/chat`. This removes the last
thing that needed a persistent local store and locks **Option A**.

## Consumer inventory (18 non-test importers) → replacement

| Area | Files | Uses PowerSync for | Replacement under Option A |
|---|---|---|---|
| Lifecycle wiring | `internal/cmd/root.go`, `internal/cmd/internal_powersync.go` | creates `Syncer`, wires uploader | delete syncer/uploader wiring; keep GraphQL services |
| App shell | `internal/app/app.go` | holds `syncer`, injects into statusbar/onboarding | remove `syncer` field + injections |
| Status bar | `internal/app/statusbar/statusbar.go`, `syncstatus/*` | sync dot + sync-error toasts | remove `syncstatus` entirely (no sync to show) |
| Sync events | `internal/app/events/sync.go` | `SyncStateChanged` etc. | remove |
| Onboarding | `internal/app/onboarding/onboarding.go`, `onboarding/sync/{model,update,view}.go` | "waiting for first sync" gate (`IsReady`) | replace with an initial GraphQL fetch / readiness check, or drop the gate |
| Uploader | `internal/upload/*` (uploader, handlers) | drains `ps_crud` → GraphQL | replace queue-drain with **synchronous GraphQL mutations** at write sites; keep handler→mutation mapping logic |
| Local DB | `internal/sqlite/*` | PowerSync-backed SQLite + generated read/write surfaces | reads → GraphQL queries; remove write surfaces + extension load |
| Schema gen | `internal/sqlite/generate/main.go`, `internal/powersync/extension/generate/main.go` | reflect PowerSync schema | delete |
| Boundary | `internal/boundary/powersync/*`, `internal/powersync/**` | the whole engine | delete |
| Test helpers | `powersynctest`, `messagelisttest` | mock syncer | delete / simplify |

## Read-path migration (CP → GraphQL queries)

Map each synced read surface to a GraphQL query. Existing services
(`internal/boundary/graphql/*_service.go`) cover some; the rich status/summary
reads need **new** operations against the control plane's current schema.

| Today (synced SQLite) | GraphQL replacement | Exists? |
|---|---|---|
| `Conversations()`, `Messages()` | `conversation_service.go`, `message_service.go` | ✅ mostly |
| `Services()` | `service_service.go` | ✅ |
| `LogEvents()`, `LogEventPolicies()` | `policy_service.go` (extend) | ⚠️ partial |
| `DatadogAccountStatuses().GetSummary()` | `account` query (canonical view) | ❌ new op |
| `ServiceStatuses()`, `LogEventStatuses()` | `services`/`logEvents` status fields | ❌ new op |
| `LogEventPolicyCategoryStatuses()` (the "Checks" surface) | `checks` query (`Check`, `CheckConnection`) | ❌ new op |
| *(new)* Issues surface | `issues` query (`Issue`, `IssueConnection`) | ❌ new op |
| *(new)* Edge instances surface | `edgeInstances` query | ❌ new op |

Action: regenerate the genqlient schema mirror
(`internal/boundary/graphql/gen/schema.graphql`) against the current control
plane — it is missing `Issue`, `Check`, `IssueConnection` — then add the queries
above. The status-bar `surfaces.go` `fetch*` functions get re-pointed from
`db.*` to the GraphQL services (with the CLAUDE.md caveat that tab update/render
paths stay non-blocking — fetch in `tea.Cmd`, which `tabpoll` already does).

## Write-path migration (ps_crud outbox → synchronous mutations)

Today: `m.db.Conversations().Create()` / `Messages().Create*()` write through
PowerSync views → captured in `ps_crud` → uploader → GraphQL mutation.

Under Option A: write sites call the GraphQL mutation directly (in a `tea.Cmd`),
keep optimistic in-memory rendering, and reconcile on success/failure. The
handler→mutation logic in `internal/upload/*_handler.go` is reusable — it just
moves from "queue consumer" to "called inline." Write sites to convert:

- `internal/app/chat/input_flow.go` (conversation + user message create)
- `internal/app/chat/usecase/{tool_loop,assistant_persistence}.go` (assistant /
  tool-result messages)
- `internal/app/chattools/setserviceenabled.go` (service patch)
- policy approve/dismiss flows (`policyHandler` equivalent)

Risk: lose the automatic retry/queue durability PowerSync provided. Mitigate
with explicit retry + error toasts (the uploader already had a retry policy we
can lift).

## Deletion inventory (once reads/writes are migrated)

- `internal/powersync/**` (engine, syncer, stream capture, crud queue)
- `internal/boundary/powersync/**` (admin API client)
- `internal/upload/**` (replaced by inline mutations)
- `internal/app/statusbar/syncstatus/**`, `internal/app/events/sync.go`
- `internal/app/onboarding/sync/**` (or repurpose as a generic loading gate)
- PowerSync-specific SQLite: extension load in `database.go`, `schema.sql`,
  `internal/powersync/extension/**`, `internal/sqlite/generate/**`
- `*_cache` read surfaces in `internal/sqlite/` and their `gen/` SQL
- Config: `PowerSyncEndpoint`, `POWERSYNC_API_TOKEN`, embedded extension binary

## Sequencing (phased, each phase shippable)

1. **Schema refresh + new queries.** Regenerate genqlient against current CP; add
   `issues`/`checks`/`account-summary`/`edgeInstances` queries + domain types.
   No deletions yet. (This is also the original "adopt new APIs" work.)
2. **Reads → GraphQL.** Re-point `surfaces.go` + status bar + onboarding reads to
   the new GraphQL services. PowerSync still running but reads no longer depend
   on synced tables. Verify parity.
3. **Writes → mutations.** Convert chat/policy/service write sites to inline
   GraphQL mutations with optimistic UI + retry. Stop relying on `ps_crud`.
4. **Remove the engine.** Delete `internal/powersync`, `internal/upload`,
   `boundary/powersync`, syncstatus, sync onboarding, PowerSync SQLite substrate,
   config, embedded extension. Drop the local DB (or swap to plain SQLite if we
   reverse the Option A decision).
5. **Cleanup.** Docs (`architecture/data-flow.md`, `domains/statusbar.md`),
   `CLAUDE.md` code-location table, dead test helpers.

## Principle: aggregation is always server-side

**Always read aggregates from the control-plane GraphQL APIs; the CLI must not
compute them client-side.** Confirmed 2026-06-08, and the control plane supports
this. This follows `CLAUDE.md` ("intelligence lives in the control plane",
"CLI is presentation only") and removes drift between CLI and webapp numbers.

Concrete implication for the current mirror work: `surfaces.go` today computes
several aggregates locally and these must be replaced by server-provided fields
in phase 2 —

- `fetchIssues`: `high = PolicyPendingCriticalCount + PolicyPendingHighCount`
  → read a server `high`/severity rollup.
- `fetchChecks`: `len(waste)+len(quality)+len(compliance)`, `categorySummary`
  pending sums, `pendingCategorySignal` high + cost sums → read `checks`
  aggregate fields.
- `fetchLogEvents`: coverage percentage `analyzed/total` → read server coverage.
- `fetchPolicies`: total fallback summing pending/approved/dismissed → read the
  server total.

## Open questions

Resolved (2026-06-08): offline **not required**; chat history **available over
GraphQL**; edge instances **in scope** for phase 1; **aggregation is always
server-side** (see principle above) — so phase 2 is a straight re-point to
server aggregates, no client-side aggregation.

None remaining that block starting phase 1.

## Testing

- Phase 2/3: snapshot/parity tests comparing GraphQL-sourced view state against
  the previous synced-SQLite values where both exist.
- Keep behavior-first tests (`operations/testing.md`); mock the GraphQL services,
  not a DB.
- Manual `/verify` of chat compose/persist + status bar after each phase.
