# Tero CLI

Presentation layer for the Tero control plane. TUI, MCP server, traditional commands.

## Read First

Start here, then pick your path:

| Doc | What You'll Learn |
|-----|-------------------|
| [docs/README.md](docs/README.md) | Manual entrypoint and reading order. |
| [docs/foundations/01-what-this-repo-is.md](docs/foundations/01-what-this-repo-is.md) | Big picture boundary and what this repository owns. |
| [docs/foundations/03-codebase-map.md](docs/foundations/03-codebase-map.md) | How the repository is organized. |
| [docs/foundations/04-data-flow.md](docs/foundations/04-data-flow.md) | How data flows between control plane, SQLite, PowerSync, and the UI. |

Then read the relevant pattern docs for the change you are making:

| Doc | When to Read |
|-----|--------------|
| [docs/patterns/architecture/tui.md](docs/patterns/architecture/tui.md) | Working on Bubble Tea models, message flow, layout, and screen composition. |
| [docs/patterns/architecture/read-models.md](docs/patterns/architecture/read-models.md) | Adding presentation-oriented local reads. |
| [docs/patterns/architecture/services.md](docs/patterns/architecture/services.md) | Working on service boundaries and choosing between local and remote implementations. |
| [docs/patterns/engineering/logging.md](docs/patterns/engineering/logging.md) | Adding or tightening logs around lifecycle, runtime, or UI behavior. |
| [docs/patterns/engineering/testing.md](docs/patterns/engineering/testing.md) | Deciding what layer should prove a behavior. |

Supporting docs:

| Doc | When to Read |
|-----|--------------|
| [docs/foundations/05-hard-rules.md](docs/foundations/05-hard-rules.md) | Repo-wide architectural guardrails. |
| [docs/meta/documentation.md](docs/meta/documentation.md) | Writing and maintaining docs in this repo. |
| [docs/meta/agent-docs.md](docs/meta/agent-docs.md) | What should live in `AGENTS.md` versus the manual. |

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
cmd/                    Composition. Binaries and high-level executable tests.
internal/interfaces/    User-facing surfaces: CLI, TUI, MCP.
internal/runtime/       Long-running coordination and deterministic progression.
internal/readmodels/    Presentation-oriented local reads.
internal/domains/       Business-shaped types and services.
internal/infrastructure/ Concrete adapters: GraphQL, SQLite, PowerSync, auth, logging.
```

## Commands

```bash
task do              # Format, lint, test - run before commits
task run             # Fast iteration
tail -f ~/.tero/environments/dev/tero.log  # Watch logs (use prd intentionally)
```
