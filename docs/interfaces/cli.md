# CLI Interface

The command surface exists for direct, scriptable operations that should not
require entering the interactive TUI runtime.

In this repo, CLI commands are adapters over existing auth/config/API
boundaries. They should orchestrate, not decide product policy.
That keeps command behavior predictable and aligned with other clients.

## How commands are wired

`internal/cmd/root.go` composes the command tree and shared dependencies.
Subcommands such as `auth`, `reset`, and `internal` each live in their own
files and receive scoped services.

This wiring pattern matters because it keeps command behavior testable and
prevents hidden global state from leaking across commands.

## What a good command implementation looks like

A strong command handler reads like a short pipeline:

1. resolve runtime/environment context,
2. validate user input and intent,
3. call service/client boundaries,
4. render result with clear success/failure semantics.

You can see this pattern clearly in `auth` flows (`internal/cmd/auth.go`) and
in internal diagnostics commands under `internal/cmd/internal_*`.
If a handler stops reading like a pipeline, it is usually a sign that concerns
should be split.

## What should not happen in command handlers

Command handlers should not become policy engines. If logic belongs to domain
rules, it should live in upstream services/control plane contracts.

They also should not mix unrelated concerns in one function (flag parsing,
network policy, formatting, local file mutation) without clear boundaries.
Small command functions are easier to reason about and much easier to test.

## Scope and environment behavior

Commands run against an explicit environment (`local`, `dev`, `prd`) through
configuration and token stores scoped to that environment.

For commands that interact with account/org data, scoping must stay explicit so
calls and local state mutations target the intended tenant context.
