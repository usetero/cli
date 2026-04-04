# TUI

The TUI is the most debt-sensitive part of this repository.

It is easy to keep adding screens and components in Bubble Tea. It is also easy
to let message ownership, routing, layout, and local state boundaries get muddy
until the whole surface becomes harder to extend than it should be. This doc is
the pattern guide for avoiding that.

## What The TUI Owns

The TUI owns the terminal interaction layer:

- screen and component composition,
- local interaction state,
- routing between active children,
- rendering,
- translating user intent into calls on runtime, read-model, or domain
  boundaries.

It does not own product truth, long-running workflow semantics, or ad hoc
storage and network behavior.

## Models Own Presentation State

A TUI model should own presentation state.

That includes things like:

- local selection state,
- input state,
- help text,
- busy or error state,
- which child is active.

What it should not quietly own is:

- cross-step product truth,
- deterministic workflow progression,
- direct database or network behavior in hot paths,
- policy that other surfaces also need to follow.

If a model starts becoming the easiest place to hide everything, that is the
first sign the boundary is drifting.

## Keep Models Narrow

A model should usually have one clear job.

This repo already leans toward splitting TUI code into focused files like
`model.go`, `render.go`, `input.go`, `busy.go`, `messages.go`, and similar.
That is a good pattern. The point is not style. The point is to keep the code
readable as screens grow.

## Do Work Through Messages

The event loop should stay honest.

If a model needs to fetch data, save state, or trigger some workflow step, that
work should happen through a `tea.Cmd` and come back into `Update` as a typed
message.

Do not do blocking database or network work directly inside synchronous TUI
paths.

That protects three things:

- responsiveness,
- testability,
- clarity about where state changes happen.

The Understanding screen at
[`internal/interfaces/tui/screens/understanding/model.go`](/Users/ben/Code/usetero/cli/internal/interfaces/tui/screens/understanding/model.go)
is a simple example of the intended shape: initialize, issue a load command,
handle the typed result, then render from state.

## Message Design Matters

Message design is an architectural concern in this repo, not just a naming
detail.

The package that emits a message should usually own that message type.

Child models should emit semantic intent messages like "submitted", "selected",
or "created", not low-level widget facts the parent has to interpret. Result
messages should carry data or errors back into the loop. They should not hide
state mutation in closures.

The useful split is:

- Bubble Tea system and input messages,
- child intent messages,
- async result messages,
- shell-level messages.

Every message should have a clear owner.

## Parents Compose; Children Behave

Parents should own composition and route selection.

Children should own their own local interaction behavior.

What parents should not do is become giant message filters that need to know
every child-specific detail in order for the tree to work. That creates brittle
UI where a new deep behavior requires edits several layers up just to keep the
message alive.

This repo already has helpers for this:

- [`internal/interfaces/tui/core/Children`](/Users/ben/Code/usetero/cli/internal/interfaces/tui/core/children.go)
- [`internal/interfaces/tui/core/Router`](/Users/ben/Code/usetero/cli/internal/interfaces/tui/core/router.go)

Use them unless there is a clear reason not to.

## Do Not Let Screens Become Workflows

A screen can initiate a workflow action without owning the workflow.

If a model starts:

- deciding progression rules,
- rebuilding product truth from multiple sources,
- accumulating too much cross-step state,
- hiding business decisions inside message handlers,

then the logic probably belongs elsewhere.

Usually that means:

- `internal/runtime/` for deterministic progression and lifecycle,
- `internal/readmodels/` for local display-oriented shaping,
- `internal/domains/` for business-shaped service behavior.

The onboarding model under
[`internal/interfaces/tui/screens/onboarding/model.go`](/Users/ben/Code/usetero/cli/internal/interfaces/tui/screens/onboarding/model.go)
is a good example because it routes user intent through the onboarding runtime
instead of making the TUI itself the owner of onboarding truth.

## Layout Is A Shell Concern First

Screens should not each invent their own shell layout rules.

The useful model for the TUI is that the shell owns regions and the screen owns
content within the body region. Header and footer placement, shell rhythm, and
global layout behavior should stay consistent across screens.

If layout starts getting fixed screen by screen with local padding hacks, the
surface quality regresses quickly.

### Intrinsic Size vs Assigned Size

When a component participates in layout, keep intrinsic size and assigned size
strictly separate.

- `PreferredHeight(width)` is intrinsic. It should come from content and local
  presentation state.
- `SetSize(width, height)` is assigned viewport. It should apply the space the
  parent chose.

Do not feed one back into the other.

In particular:

- do not store an assigned viewport height and later return it from
  `PreferredHeight`,
- do not measure a rendered viewport-style child to discover its preferred
  height,
- if a component truly needs both values, store them in separate clearly named
  fields.

The parent owns layout. The child owns its intrinsic size.

## Shared Presentation Should Stay Shared

This repo already has shared presentation primitives under
[`internal/interfaces/tui/ui/present`](/Users/ben/Code/usetero/cli/internal/interfaces/tui/ui/present)
and semantic theme tokens under
[`internal/interfaces/tui/ui/theme`](/Users/ben/Code/usetero/cli/internal/interfaces/tui/ui/theme).

Use those shared surfaces and tokens instead of inventing one-off visual rules
inside each screen. If the same pattern shows up in multiple places, promote it
into a shared primitive rather than duplicating it.

## Render From State

Rendering should be a projection of model state, not a second logic engine.

When render code starts carrying too much hidden conditional behavior, it is
usually compensating for a state boundary that is already wrong.

## What To Test

The best TUI tests in this repo are the ones that protect boundaries:

- route ownership,
- message forwarding,
- busy and error transitions,
- typed-message updates,
- rendering contracts that matter for navigation and comprehension.

Do not default to exhaustive snapshots. Prefer tests that catch the kinds of
drift that make the TUI harder to change safely.
