# Codebase Map

This chapter is the practical map of the repository.

The code is organized around ownership. Once you understand the main
directories, it becomes much easier to place new code correctly and much easier
to tell when something is drifting into the wrong layer.

## Top-Level Shape

The main structure is:

- `cmd/`
- `internal/interfaces/`
- `internal/runtime/`
- `internal/readmodels/`
- `internal/domains/`
- `internal/infrastructure/`

That shape is not accidental. Each of those areas has a different job.

## `cmd/`

[`cmd/`](/Users/ben/Code/usetero/cli/cmd) is the composition layer.

This is where binaries start and where dependencies are wired together. It is
also where the highest-level executable tests live, because those tests are
proving the binary surface rather than an isolated package.

Keep `cmd/` thin. It should assemble the app, not become the place where
business or workflow semantics are invented.

## `internal/interfaces/`

[`internal/interfaces/`](/Users/ben/Code/usetero/cli/internal/interfaces) is
where the user-facing surfaces live.

That includes:

- [`internal/interfaces/cli`](/Users/ben/Code/usetero/cli/internal/interfaces/cli)
- [`internal/interfaces/tui`](/Users/ben/Code/usetero/cli/internal/interfaces/tui)
- [`internal/interfaces/mcp`](/Users/ben/Code/usetero/cli/internal/interfaces/mcp)

These packages translate user intent into calls on the right boundaries and then
render the result appropriately for the surface.

The key point is that they are interface code. They should not grow into a
second product logic layer.

## `internal/runtime/`

[`internal/runtime/`](/Users/ben/Code/usetero/cli/internal/runtime) owns
coordination and state progression over time.

This is where the repository puts logic that should not live directly inside TUI
models or command handlers, especially when the logic is about lifecycle,
projection, or long-running account-scoped behavior.

Two important examples are:

- onboarding under
  [`internal/runtime/onboarding`](/Users/ben/Code/usetero/cli/internal/runtime/onboarding)
- account runtime under
  [`internal/runtime/account`](/Users/ben/Code/usetero/cli/internal/runtime/account)

## `internal/readmodels/`

[`internal/readmodels/`](/Users/ben/Code/usetero/cli/internal/readmodels)
exists to keep presentation code from becoming database-heavy.

Read models shape local synced data into forms that are useful for rendering and
interaction. They are a presentation-oriented boundary, not a general domain
abstraction.

The Understanding read model is the clearest current example of this direction.

## `internal/domains/`

[`internal/domains/`](/Users/ben/Code/usetero/cli/internal/domains) holds
business-shaped types and services.

This is where the repo models concepts like tenancy, catalog, integrations,
identity, preferences, change, and chat. Some of those domains expose both
local and remote implementations because the CLI genuinely needs both paths.

The important pattern is that the service boundary stays aligned with the
business entity, not with transport details.

## `internal/infrastructure/`

[`internal/infrastructure/`](/Users/ben/Code/usetero/cli/internal/infrastructure)
contains the concrete capabilities underneath everything else:

- control-plane GraphQL access,
- SQLite,
- PowerSync,
- logging,
- auth,
- preferences,
- chat infrastructure.

Infrastructure should stay concrete. It should provide capabilities to the rest
of the app, not quietly take over ownership of product behavior.

## A Good Placement Test

When you are deciding where code belongs, a simple test helps:

- if it is about starting binaries or wiring dependencies, it belongs in `cmd/`
- if it is about a user-facing surface, it belongs in `internal/interfaces/`
- if it is about state progression or lifecycle over time, it belongs in
  `internal/runtime/`
- if it is about shaping local reads for presentation, it belongs in
  `internal/readmodels/`
- if it is about a business-shaped service or concept, it belongs in
  `internal/domains/`
- if it is about a concrete technical capability, it belongs in
  `internal/infrastructure/`

That test is not perfect, but it is usually enough to catch obvious placement
mistakes.
