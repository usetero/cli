# Linting

Linting is organized as one repo-owned entrypoint with focused rule modules.

## Entry Point

- `scripts/lint/cmd/lint/main.go`

Run all lint suites:

```bash
go run ./scripts/lint/cmd/lint
```

Run one suite:

```bash
go run ./scripts/lint/cmd/lint --suite analyzers
```

Supported suites:

- `all`
- `repo`
- `analyzers`
- `scripts`

## Rule Modules

- `scripts/lint/analyzers/`: AST-aware Go analyzers for repo doctrine.
- `scripts/lint/check-*.sh`: narrow shell checks that still encode repo guardrails.

The current analyzer set covers:

- `archdeps`: high-level package layer boundaries
- `ctxfirst`: `context.Context` comes first when present
- `teacmdpurity`: no model mutation inside `tea.Cmd` closures

## Philosophy

Linting in this repo exists to protect architecture, ownership, and TUI runtime
discipline. Generic style belongs to `golangci-lint`; repo doctrine belongs
here.
