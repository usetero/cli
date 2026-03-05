# Domain

This file mirrors `AGENTS.md` for tools that still load `CLAUDE.md`.
Authoritative source: `AGENTS.md`.

Shared domain types for all interfaces (TUI, CLI, MCP, chat tools).

## Rules

1. `internal/domain` must not import presentation packages.
2. Parse raw DB/API payloads once in constructors/parsers, then pass rich types downstream.
3. Prefer typed enums over raw strings when value sets are known.
4. Derived cross-interface values belong on domain models, not UI renderers.

## Policy Modeling Pattern

Policy data uses a raw-plus-rich split:

1. Raw DTO mirrors storage format.
2. Rich model exposes typed fields and parsed analysis.
3. Parser bridges raw -> rich in one place.

Do not duplicate parsing in callers.

## Analysis Extensions

When adding a new policy analysis category:

1. Add/update analysis struct in `policy_analysis_*.go`.
2. Implement/override required `PolicyAnalysis` methods.
3. Register parser dispatch in `policy_analysis.go`.
