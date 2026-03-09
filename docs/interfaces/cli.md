# CLI Interface

The command surface exists for direct, scriptable operations that should not
require entering the interactive TUI runtime.

In this repo, CLI commands are adapters over existing auth/config/API
boundaries. They should orchestrate, not decide product policy.
That keeps command behavior predictable and aligned with other clients.

## How commands are wired

[`internal/interfaces/cli/execute.go`](../../internal/interfaces/cli/execute.go)
owns config resolution and surface selection.

Individual command entrypoints then compose their own narrow dependencies under
[`internal/interfaces/cli`](../../internal/interfaces/cli).

This wiring pattern matters because it keeps command behavior testable and
prevents hidden global state from leaking across commands.

## What a good command implementation looks like

A strong command handler reads like a short pipeline:

1. resolve runtime/environment context,
2. validate user input and intent,
3. call service/client boundaries,
4. render result with clear success/failure semantics.

The current surface is intentionally small, but the same rule applies as more
commands are added. If a handler stops reading like a pipeline, it is usually a
sign that concerns should be split.

## What should not happen in command handlers

Command handlers should not become policy engines. If logic belongs to domain
rules, it should live in upstream services/control plane contracts.

They also should not mix unrelated concerns in one function (flag parsing,
network policy, formatting, local file mutation) without clear boundaries.
Small command functions are easier to reason about and much easier to test.

## Scope and environment behavior

Commands run against an explicit environment (`local`, `dev`, `prd`) through
configuration and token stores scoped to that environment.

Service environment variables should store origins, not transport-specific
endpoints. For example, `TERO_API_ORIGIN` should be `https://api.usetero.dev`,
not `https://api.usetero.dev/graphql`. Clients append their own fixed paths.

For commands that interact with account/org data, scoping must stay explicit so
calls and local state mutations target the intended tenant context.
