# Tero CLI

Presentation layer for the Tero control plane. TUI, MCP server, traditional commands.

## Read First

| Working on | Read |
|------------|------|
| Anything | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) |
| Tests | [docs/TESTING.md](docs/TESTING.md) |
| Logs | [docs/LOGGING.md](docs/LOGGING.md) |
| Product context | [product.md](../knowledge-base/product.md) |

## Rules

1. **CLI is presentation only.** Intelligence lives in the control plane.
2. **Control plane is source of truth.** Local data is synced, never owned.
3. **Dependencies point inward.** Services depend on interfaces, not implementations.
4. **Composition happens in `cmd/`.** Wire implementations to interfaces there, nowhere else.
5. **Conventional commits.** `feat:`, `fix:`, `docs:`, `refactor:`, etc.

## Code Location

```
Wiring dependencies?         → cmd/
Domain logic?                → internal/{domain}/
Implementation details?      → internal/{impl}/
TUI presentation?            → internal/tui/
```

## Commands

```bash
task do      # Format, lint, test - run before commits
task run     # Fast iteration
```
