# TUI Interface

The TUI is split into a thin app shell and focused screen models.

Current ownership:

- `internal/interfaces/tui/root/model.go`:
  app shell only (global quit, framing, top-level composition).
- `internal/interfaces/tui/screens/onboarding/model.go`:
  onboarding flow orchestration (runtime calls + step routing).
- `internal/interfaces/tui/screens/onboarding/*`:
  focused UI models for each step (role, organization/select, organization/create, account/select).

## Core Pattern

1. Root composes, screens decide.
2. Runtime calls stay in the orchestration model (`screens/onboarding/model.go`), not in leaf UI models.
3. Leaf models own local UI state only.
4. All async work returns typed `tea.Msg` contracts.

## Message Contracts

Message contracts are owned by the package that emits them.

Examples:

- `screens/onboarding/role/submitted.go` -> `SubmittedMsg`
- `screens/onboarding/organization/select/selected.go` -> `SelectedMsg`
- `screens/onboarding/organization/create/created.go` -> `CreatedMsg`
- `screens/onboarding/account/select/selected.go` -> `SelectedMsg`

Guidelines:

- Prefer event/fact names (`Submitted`, `Selected`, `Created`).
- Keep payloads minimal and strongly typed.
- Avoid shared catch-all message files at root.

## Step Structure

For multi-mode entities (like organization), split by intent:

- `organization/select`
- `organization/create`

This keeps each model small and keeps update/view logic obvious.

## Runtime Progression Contract

The onboarding screen model uses runtime state as source of truth:

1. `Init` loads state via `Runtime.State`.
2. User action emits message from leaf model.
3. Orchestrator calls runtime method (`SetRole`, `SelectOrganization`, `CreateOrganization`, `SelectAccount`).
4. Returned state determines next route.

Auto-selection behavior (single option, valid preference match) remains in runtime/domain logic, not in leaf UI models.

## Bubble Tea Rules

- No network/DB work in `View`.
- Keep `Update` fast and deterministic.
- Side effects only in `tea.Cmd`.
- Re-enter through typed messages.

## Input Policy

- Root view owns global terminal policy:
  - `AltScreen` enabled,
  - `WindowTitle` set,
  - `MouseMode` defaults to `MouseModeNone` unless a screen explicitly needs mouse interaction.
- Program startup applies a global input filter (`internal/interfaces/tui/filter`) to throttle noisy mouse motion/wheel bursts.
- Enable mouse modes only in screens that implement click/hover behavior.

## Shell Contract

- The shell has three fixed slots: header, body, footer.
- Header and footer are pinned; body expands to consume remaining viewport height.
- Global shell chrome does not draw a universal panel around page content.
- Emphasized content (errors, alerts) uses reusable card chrome helpers in the screen body.

## Rendering Boundaries

- `chrome/*` is for presentational helpers only (shell, card, layout wrappers).
- `components/*` is for interactive Bubble Tea models (input, list, progress, status, help).
- If it has no `Update`/`Init` state and only renders strings, keep it in `chrome`.

## Model Contract

Nested screen models should implement `internal/interfaces/tui/screen.Model`:

- `tea.Model` (`Init`, `Update`, `View`)
- `SetSize(width, height int)`
- `ShortHelp() []key.Binding`

This contract is enforced with compile-time assertions in onboarding leaf and flow models.

### Routing Discipline

- Parent models handle only:
  - global keys
  - layout (`tea.WindowSizeMsg`)
  - parent-owned typed intent messages
- Otherwise, parents forward messages to the active child model.
- `ShortHelp` cascades from root to active child; key bindings are defined in the same model that handles them.
- Router child IDs are typed enums per parent model, not raw string literals.

## Lint Guardrails

Architecture lint now enforces the child-router contract for router-backed parent models:

- parent updates must forward through `router.Forward(msg)`,
- parent state transitions must explicitly activate/deactivate children (`ActivateOnly`, `SetActive`, `ClearActive`),
- `ShortHelp` must cascade through `router.ShortHelp()`,
- `tea.Cmd` closures must emit messages only (no parent state mutation inside closures).

Run with:

```bash
task lint:architecture
```
