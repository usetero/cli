# Docs Index

This directory is the source of truth for engineering decisions and workflows in this repo.

## Start Here

1. [ARCHITECTURE.md](ARCHITECTURE.md): system boundaries and dependency rules.
2. [DATA.md](DATA.md): how data moves and where truth lives.
3. [TEA.md](TEA.md): Bubble Tea component architecture and UI state patterns.
4. [TESTING.md](TESTING.md): test strategy, invariants, and required checks.

Then pick interface-specific docs:

1. [CLI.md](CLI.md)
2. [MCP.md](MCP.md)

Operational docs:

1. [LOGGING.md](LOGGING.md)
2. [OBSERVABILITY.md](OBSERVABILITY.md)
3. [TOASTS.md](TOASTS.md)

Chat-specific test planning:

1. [CHAT_TEST_AUDIT.md](CHAT_TEST_AUDIT.md)

## Documentation Standards

1. Keep docs stable and policy-oriented; avoid restating code details that drift quickly.
2. Put architecture decisions and invariants here; keep implementation mechanics in code/tests.
3. Prefer one source of truth per topic. Link, do not duplicate.
4. Update docs in the same change when behavior or policy changes.

## Ownership Model

1. `ARCHITECTURE.md`, `DATA.md`, `TESTING.md`, `TEA.md` are foundational and should change deliberately.
2. Interface docs (`CLI.md`, `MCP.md`) describe contracts and expectations, not internals.
3. Operational docs (`LOGGING.md`, `OBSERVABILITY.md`, `TOASTS.md`) document runtime behavior and oncall workflows.
