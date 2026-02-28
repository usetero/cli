# Debug Workflows

Standard debugging workflows for Tero CLI. These workflows are task-driven and environment-safe.

## Principles

1. Use `task` as the workflow entrypoint.
2. Use `dev` (or `local`) by default.
3. Do not run iterative debugging against `prd` unless explicitly required.
4. Keep debug commands repeatable and composable.

## Environment Policy

Default debug environment is `dev`.

1. `task debug:*` defaults to `TERO_ENV=dev`.
2. `prd` is blocked unless `ALLOW_PRD=1` is explicitly set.

Example:

```bash
task debug:ui:logs:slow
task debug:ui:logs:slow TERO_ENV=prd ALLOW_PRD=1
```

## Command Map

List available debug workflows:

```bash
task debug:list
```

UI:

```bash
task debug:ui:run
task debug:ui:smoke
task debug:ui:logs:slow
task debug:ui:logs:all
```

Onboarding:

```bash
task debug:onboarding:trace
```

## Slow UI Diagnostics

Slow-loop telemetry logs:

1. `slow app update`
2. `slow app render`

Each log includes context to attribute slowness:

1. `duration_ms`
2. `state`
3. `msg_type` (for updates)
4. `drawer_open`
5. `drawer_tab`
6. overlay state (`palette_open`, `quit_dialog_open`)

Recommended loop:

1. Run app with `task debug:ui:run`.
2. In a second terminal, run `task debug:ui:logs:slow`.
3. Reproduce lag once.
4. Use emitted context to identify whether slowness is in update or render.

## Naming Convention

Use `debug:<area>:<action>`:

1. `debug:ui:*` for loop/render/input workflows.
2. `debug:onboarding:*` for onboarding gates and transitions.
3. `debug:sync:*` for sync/upload behavior.
4. `debug:data:*` for SQLite/data snapshots.

Add new debug tasks in `taskfiles/debug.yml`.

