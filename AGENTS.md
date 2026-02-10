# Tero CLI

Presentation layer for the Tero control plane. TUI, MCP server, traditional commands.

## Read First

Start here, then pick your path:

| Doc | What You'll Learn |
|-----|-------------------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Big picture. Three interfaces, two data patterns, code structure. |
| [DATA.md](docs/DATA.md) | How data flows. Local vs remote, SQLite, PowerSync. |

Then pick your interface:

| Doc | When to Read |
|-----|--------------|
| [TEA.md](docs/TEA.md) | Working on the terminal UI. Bubbletea patterns, models, layout. |
| [MCP.md](docs/MCP.md) | Working on the MCP server. (Not yet implemented.) |
| [CLI.md](docs/CLI.md) | Adding CLI commands. Direct API access. |

Supporting docs:

| Doc | When to Read |
|-----|--------------|
| [TESTING.md](docs/TESTING.md) | Writing tests. |
| [LOGGING.md](docs/LOGGING.md) | Writing logs. |

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
internal/tui/        TUI. Bubbletea models, pages, chat.
internal/cmd/        CLI commands. Direct API calls.
internal/chat/       Chat client. Streaming, accumulation.
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
tail -f ~/.tero/environments/prd/tero.log  # Watch logs
```
