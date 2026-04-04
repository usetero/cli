# Code Shape

This repository should feel boring to work in.

That is not an aesthetic preference. It is a practical one. The codebase already
has enough moving parts: TUI message flow, runtime coordination, local SQLite,
PowerSync, and control-plane calls. The way it stays manageable is by keeping
ownership obvious and shaping code so a reader can tell quickly where something
belongs.

This doc is here to make that standard explicit.

## Optimize For Placement First

The first question when writing code here is usually not "how do I abstract
this?"

It is:

where does this belong?

Most bad code shape in this repo is really a placement mistake.

The main ownership choices are:

- `cmd/` for composition,
- `internal/interfaces/` for user-facing surfaces,
- `internal/runtime/` for coordination and lifecycle over time,
- `internal/readmodels/` for local presentation-shaped reads,
- `internal/domains/` for business-shaped services,
- `internal/infrastructure/` for concrete capabilities.

If the placement is wrong, the shape will usually feel wrong too.

## Keep Files Small And Single-Purpose

A file should usually have one clear reason to exist.

When a concept grows, split it into a few obvious files instead of letting one
file absorb everything. The TUI already follows this pattern heavily with files
like:

- `model.go`
- `render.go`
- `input.go`
- `busy.go`
- `messages.go`

That pattern is useful because it keeps a model readable as it grows. The same
instinct should show up outside the TUI too. If a file starts mixing mapping,
validation, orchestration, and rendering or transport logic together, it is
almost certainly taking on too much.

## Name Files For The Concept, Not The Junk Drawer

Avoid grab-bag files and vague buckets.

Names like `helpers.go`, `utils.go`, or a generic `types.go` are almost always a
signal that the code has not been split cleanly enough yet. Prefer files whose
name tells the reader what part of the concept they are looking at.

The repo already has a good instinct for this in many places. Keep leaning into
that.

## Keep Logic Close To The Owner

If code exists for one screen, one service, or one read model, keep it near that
owner unless there is strong and proven reuse.

This repository gets more value from obvious local ownership than from broad
shared helper layers. A small amount of duplication is often cheaper than an
abstraction that makes it harder to tell where the real behavior lives.

This is especially true for:

- local query shape,
- screen-specific TUI logic,
- read-model shaping,
- adapter mapping code.

## Queries Belong Near The Caller

This repository deliberately keeps SQL close to the package that owns the read.

You can see that in domain packages such as:

- [internal/domains/catalog/db/log_events.sql](/Users/ben/Code/usetero/cli/internal/domains/catalog/db/log_events.sql)
- [internal/domains/tenancy/db/workspaces.sql](/Users/ben/Code/usetero/cli/internal/domains/tenancy/db/workspaces.sql)

and in read-model packages such as:

- [internal/readmodels/understanding/db/snapshot.sql](/Users/ben/Code/usetero/cli/internal/readmodels/understanding/db/snapshot.sql)
- [internal/readmodels/understanding/db/event_detail.sql](/Users/ben/Code/usetero/cli/internal/readmodels/understanding/db/event_detail.sql)

That is not an accident and it should continue.

The repo is not trying to force every SQLite read through one generic ORM or one
central repository layer. Many queries here exist for one specific owner. Keep
them local so the code remains direct and easy to trace.

## Service Boundaries Should Follow Entities

Services should be aligned with business concepts, not transport details.

That means the stable thing is the domain concept. Local and remote
implementations are secondary. If a service boundary starts being defined mostly
by whether it talks to GraphQL or SQLite, the design is probably drifting.

This is part of why local and remote implementations live under the same domain
concept instead of being split into unrelated transport-specific package shapes.

## Constructors Should Be Explicit

Constructors should make required dependencies obvious and fail early when the
requirements are not met.

This repo already leans on direct constructor injection and early validation.
That is a good fit. It keeps composition inspectable and makes missing
dependencies fail loudly instead of producing strange nil behavior later.

Prefer:

- required dependencies passed directly,
- immediate validation of required inputs,
- fields named for what they hold.

## Do Not Abstract Preemptively

This repo should prefer direct code until reuse is real.

Do not create broad helper layers, wrapper packages, or reusable abstractions
just because two pieces of code look similar. Add the abstraction when it makes
the code easier to understand, not just technically more unified.

The most common failure mode here is an abstraction that removes a bit of local
duplication but makes ownership harder to see. That is a bad trade in this
codebase.

## Make It Easy To Follow A Call Path

When someone opens a file, they should be able to answer quickly:

- what layer am I in,
- what concept does this file own,
- what does it call next,
- what does it definitely not own.

That is the standard to aim for.

If the reader has to jump through several helper layers or guess where the real
work is happening, the code shape is probably wrong.

## What Drift Looks Like

Code shape is drifting when:

- files turn into mixed-responsibility grab bags,
- service boundaries start following transport rather than the entity,
- local query logic gets pulled into generic shared layers,
- helper packages grow without clear ownership,
- interfaces start hiding runtime or domain logic,
- runtime starts accumulating presentation detail,
- code becomes harder to place than to write.

The right fix is usually to restore clear ownership, not to add one more helper.
