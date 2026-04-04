# Read Models

Read models exist to keep presentation code clean.

They are the boundary between local synced data and the surface that needs to
render or navigate that data. This repo uses them because the TUI should not
have to know how to query SQLite directly just to render one screen well.

## What A Read Model Is

A read model is a read-only, presentation-oriented API over local data.

It exists to answer questions like:

- what should this screen render,
- what shape should that data have,
- what local view of the account does the interface need right now.

The Understanding read model under
[`internal/readmodels/understanding`](/Users/ben/Code/usetero/cli/internal/readmodels/understanding)
is the clearest current example.

## Why They Exist

Without read models, the usual fallback is to let the TUI reach directly into
SQLite or to keep pushing more and more query-specific logic into domain
services.

Both of those options create debt.

If the TUI owns the query layer, presentation code becomes data-access code.
If domain services absorb every display-shaped read, the domain layer gets wider
and less coherent than it should be.

Read models are the middle path that keeps both sides cleaner.

## What Good Read Models Do

A good read model:

- exposes a small API shaped around a surface,
- hides raw query plumbing from the caller,
- returns data in a form that is easy to render or interact with,
- stays read-only.

The point is not abstraction for its own sake. The point is to give a screen or
component a stable local view without dragging storage details into the UI.

## What Read Models Should Not Become

Read models should not become:

- generic repositories for unrelated reads,
- write owners,
- alternate domain services,
- a shadow authority for product truth.

If a read-model package starts looking like a general-purpose service layer, the
boundary is drifting.

## Read Models Versus Domain Services

The distinction is simple:

- domain services are shaped around business entities and operations,
- read models are shaped around presentation and local exploration.

Those two things may touch the same underlying tables. They still have different
jobs. Preserving that distinction keeps the layers easier to reason about.

## Keep Queries Close To The Owner

Read-model queries should live close to the read model that owns them.

This repo gets more value from obvious local query ownership than from a large
central query layer. Most presentation reads are specific. Keeping them close to
the owner makes the intent clearer and avoids building one generic read API that
has to serve every use case badly.

## When To Add One

Add a read model when all of these are true:

- the need is primarily presentation-oriented,
- the data is local or naturally modeled as a local view,
- putting the query logic directly in the interface would make that surface too
  database-aware.

Do not add one just because a package feels crowded. Add one when it gives the
presentation layer a better boundary.
