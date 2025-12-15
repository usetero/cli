# AI Agent Bootloader

Tero CLI - presentation layer for the Tero control plane. TUI, MCP server, traditional commands.

## Before Writing Code

| Working on... | Read first |
|--------------|-----------|
| Anything | `docs/ARCHITECTURE.md` |
| TUI | `docs/TUI.md` |
| Tests | `docs/TESTING.md` |
| Design decisions | `docs/DESIGN.md` |

The docs are short. Don't skip them.

## Hard Rules

- **CLI is presentation only.** Intelligence lives in the control plane.
- **Control plane is source of truth.** CLI caches for display, never owns data.
- **Dependencies point inward.** Services depend on interfaces, not implementations.
- **Composition happens in `cmd/`.** That's where you wire implementations to interfaces.
- **Conventional commits.** `feat:`, `fix:`, `docs:`, `refactor:`, etc.

## Workflow

```bash
task do      # Format, lint, test - run before commits
task run     # Fast iteration
```

## Decision Trees

### Where does this code go?

```
Wiring dependencies?                              → cmd/
Domain logic (auth, preferences)?                 → internal/{domain}/
Concrete implementation (keyring, workos, config)?→ internal/{impl}/
TUI presentation?                                 → internal/tui/
```

### Interface or concrete type?

```
Used as a dependency by a service?      → Interface (defined by consumer)
Platform-specific or external API?      → Implementation (implements interface)
```

## When Stuck

Read the docs. They answer most questions.
