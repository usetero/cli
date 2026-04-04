# Hard Rules

This chapter is the short list of rules that matter enough to write down
explicitly.

They are not style preferences. They are the constraints that keep the CLI code
from turning into a mixture of presentation logic, product logic, and local
state plumbing with no clear ownership.

## The CLI Is Presentation

This repository is the presentation layer for the control plane.

Interfaces can orchestrate and render. They should not become independent
product policy engines.

## The Control Plane Is The Source Of Truth

Local SQLite state, PowerSync, and read models exist to support the interface.
They do not replace the control plane as the authority for the product.

## Composition Happens In `cmd/`

Keep wiring and top-level assembly in `cmd/` and obvious startup points. Do not
scatter dependency assembly through feature packages.

## TUI Models Own Presentation State

Bubble Tea models should focus on presentation state, input handling, message
handling, and rendering.

If a model starts owning too much progression logic, direct data access, or
hidden business truth, that is usually a sign the logic belongs elsewhere.

## Nothing Blocking Happens In Hot UI Paths

Network and database work should happen through commands and re-enter the event
loop through typed messages. Do not block synchronous TUI hot paths.

## Parents Should Compose, Not Filter Everything

Parent models should own composition and route selection. They should not become
fragile message filters that need to know every downstream behavior in detail.

Helpers like
[`internal/interfaces/tui/core/Children`](/Users/ben/Code/usetero/cli/internal/interfaces/tui/core/children.go)
and
[`internal/interfaces/tui/core/Router`](/Users/ben/Code/usetero/cli/internal/interfaces/tui/core/router.go)
exist to keep that delegation consistent.

## Read Models Shape Presentation-Oriented Reads

Read models are there to keep presentation code clean. They are not generic
repositories or alternate domain authorities.

## Queries Belong Close To The Owner

If a query exists to support one specific screen, read model, or local service,
keep it close to that owner unless there is a strong reason not to.

## Keep The Code Boring

This repository benefits from obvious ownership, small focused files, and direct
code paths. Prefer clarity over cleverness.
