# Documentation

This directory is a working map for engineers changing the CLI. It is organized
by the question you are trying to answer, not by package names.

If you are new to the repository, start here and then follow one of the reading
paths below.

## What This Repo Is

The CLI is a presentation runtime for the Tero control plane. It renders the
interactive TUI (`internal/app`), exposes direct command surfaces
(`internal/cmd`), and keeps terminal interactions responsive through a local
SQLite projection backed by PowerSync.

Business truth still lives in the control plane. The CLI should present and
orchestrate that truth, not redefine it.

## How This Docs Tree Is Organized

- `architecture/`: system shape, boundaries, and non-negotiable flow rules.
- `domains/`: ownership and invariants for concrete product flows (for example onboarding).
- `interfaces/`: contracts by user surface (`tui`, `cli`, `mcp`).
- `operations/`: practical debugging, logging, observability, and testing workflows.
- `reference/`: stable conventions referenced from multiple docs.
- `specs/`: focused audits and temporary design/test artifacts.
- `meta/`: standards for writing/maintaining documentation in this repo.

## Read By Intent

### I need the big picture before I touch code

- [architecture/README.md](architecture/README.md)
- [architecture/system-overview.md](architecture/system-overview.md)
- [architecture/data-flow.md](architecture/data-flow.md)
- [architecture/ui-architecture.md](architecture/ui-architecture.md)

### I am changing onboarding/bootstrap behavior

- [domains/onboarding-bootstrap.md](domains/onboarding-bootstrap.md)
- [interfaces/tui.md](interfaces/tui.md)
- [operations/testing.md](operations/testing.md)
- [operations/debugging.md](operations/debugging.md)

### I am changing chat behavior or lifecycle semantics

- [domains/chat.md](domains/chat.md)
- [interfaces/tui.md](interfaces/tui.md)
- [operations/testing.md](operations/testing.md)
- [specs/chat-test-audit.md](specs/chat-test-audit.md)

### I am adding or modifying a CLI command

- [interfaces/cli.md](interfaces/cli.md)
- [architecture/system-overview.md](architecture/system-overview.md)
- [operations/testing.md](operations/testing.md)

### I am debugging runtime issues

- [operations/debugging.md](operations/debugging.md)
- [operations/observability.md](operations/observability.md)
- [operations/logging.md](operations/logging.md)
- [reference/toasts.md](reference/toasts.md)

## Code Entry Points (Quick Map)

Use these paths when docs and code should be read together:

- `cmd/tero/main.go`: executable entrypoint.
- `internal/cmd/root.go`: dependency composition and app bootstrap wiring.
- `internal/app/app.go`: root Bubble Tea model and high-level state orchestration.
- `internal/app/onboarding/`: onboarding orchestrator and gate steps.
- `internal/core/bootstrap/`: deterministic onboarding transition engine.
- `internal/chat/`: chat protocol/reducer contracts.
- `internal/powersync/`: sync lifecycle and projection pipeline.

## Documentation Quality Bar

A doc in this repo is good if it helps an engineer build the right mental model,
predict non-happy-path behavior, understand what must not be broken, and find
the right code entrypoints quickly.

If a page is only a checklist of files/commands, it is incomplete.
