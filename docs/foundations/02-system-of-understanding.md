# System Of Understanding

The CLI architecture makes more sense once you understand the product idea
behind it.

Tero is not trying to be another system of record. It is trying to give users a
system of understanding: a way to move through production telemetry and related
state in a form that is inspectable, queryable, and useful for real decisions.

That matters here because it explains why the CLI needs both remote access and a
local synced runtime.

## What That Means In Practice

The product is not only about storing entities or showing raw provider data.

It is about helping a user answer questions like:

- what services exist,
- what events matter,
- what does the system know about them,
- what evidence supports that understanding,
- what should the user do next.

That is a different job from simply exposing an API response in a terminal.

## Why Local State Exists

If the CLI only made remote requests, it could still work, but it would be much
more limited. The terminal experience would be slower, more brittle, and less
natural for exploration.

Local SQLite plus PowerSync gives the CLI something important: a fast local
projection of account-scoped state that the interface can move through without
waiting on the network for every interaction.

That matters especially for the TUI and for screens like Understanding, where
the whole point is to let a user move around a local view of the system quickly.

## What Local State Is For

Local state exists to improve:

- responsiveness,
- navigation,
- inspection,
- transparency.

It lets the CLI show more of the system at once and lets the user move through
that view more directly.

## What Local State Is Not For

Local state is not there to replace the control plane.

The control plane remains authoritative. The local runtime exists so the CLI can
work with that truth efficiently. The repo should be comfortable building local
projections and local read paths, but it should stay disciplined about not
treating them like a separate source of product truth.

## Why This Changes The Code

Once you accept that split, a few design choices in the repository stop looking
strange:

- there are both local and remote service implementations in some domains,
- read models exist to shape local data for presentation,
- bootstrap flows are different from steady-state runtime flows,
- queries live close to the caller instead of being forced through one generic
  data layer.

Those are not random preferences. They come from the product needing both
authority and local speed.

## A Useful Way To Read The Repo

When you are reading code in this repository, it helps to keep asking:

is this code deciding truth, or is it helping the CLI present and explore truth?

That question is useful in almost every layer, especially in the TUI, runtime,
read-model, and local-service code.
