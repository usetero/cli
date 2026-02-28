# Tero CLI

Presentation layer for the Tero control plane. TUI, MCP server, traditional commands.

## Read First

Start here, then pick your path:

| Doc | What You'll Learn |
|-----|-------------------|
| [architecture/system-overview.md](docs/architecture/system-overview.md) | Big picture boundary, ownership, and runtime lifecycle. |
| [architecture/data-flow.md](docs/architecture/data-flow.md) | How data flows between GraphQL, SQLite, PowerSync, and uploader. |

Then pick your interface:

| Doc | When to Read |
|-----|--------------|
| [interfaces/tui.md](docs/interfaces/tui.md) | Working on the terminal UI runtime and Bubble Tea boundaries. |
| [interfaces/mcp.md](docs/interfaces/mcp.md) | MCP adapter constraints. (Planned surface.) |
| [interfaces/cli.md](docs/interfaces/cli.md) | Adding CLI commands and keeping handlers thin. |

Supporting docs:

| Doc | When to Read |
|-----|--------------|
| [operations/testing.md](docs/operations/testing.md) | Writing behavior-first tests and using test workflows. |
| [operations/logging.md](docs/operations/logging.md) | Writing logs that are useful in production diagnosis. |
| [operations/debugging.md](docs/operations/debugging.md) | Running debug workflows safely (`dev` by default). |

## Agent Rules

- **No sub-agents.** Never use the Task tool to spawn sub-agents. Do all work directly in the main context — read files, search, edit, run commands yourself. Background bash commands are fine.

## Rules

1. **CLI is presentation only.** Intelligence lives in the control plane.
2. **Control plane is source of truth.** Local data is synced (TUI/MCP) or queried (CLI), never owned.
3. **Dependencies point inward.** Services depend on interfaces, not implementations.
4. **Composition happens in `cmd/`.** Wire implementations to interfaces there, nowhere else.
5. **Conventional commits.** Commit prefixes drive release-please version bumps and changelogs. Think from the **user's perspective** — would they notice this change?

   | Prefix | When to use | Version bump |
   |--------|-------------|--------------|
   | `feat:` | User-visible new functionality. New UI element, new command, new capability. | Minor |
   | `fix:` | User-visible bug fix. Something was broken, now it works. | Patch |
   | `refactor:` | Internal restructuring. No user-visible change. | None |
   | `chore:` | Build, deps, CI, tooling. No user-visible change. | None |
   | `docs:` | Documentation only. | None |
   | `test:` | Test only. No production code change. | None |

   **The test:** If you'd put it in release notes for users, it's `feat:` or `fix:`. If only developers care, it's `refactor:`, `chore:`, or `test:`.

## Code Location

```
cmd/                 Wiring. Creates implementations, injects dependencies.
internal/app/        TUI. Bubble Tea models, pages, onboarding, chat.
internal/cmd/        CLI commands. Direct API calls.
internal/api/chatclient/ Chat client. Streaming, accumulation.
internal/core/chat/  Pure chat lifecycle/session policy.
internal/api/        GraphQL client. Control plane CRUD.
internal/sqlite/     Local database.
internal/powersync/  Sync engine.
internal/domain/     Shared types.
internal/auth/       Authentication.
```

## Commands

```bash
task do              # Format, lint, test - run before commits
task run             # Fast iteration
tail -f ~/.tero/environments/dev/tero.log  # Watch logs (use prd intentionally)
```
