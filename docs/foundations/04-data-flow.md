# Data Flow

The CLI uses different data paths in different phases of the product. That is
intentional.

If you flatten all of those flows into one mental model, the code can look
inconsistent. If you keep the phases separate, it is much easier to understand
why the repository is shaped the way it is.

## The Main Rule

Remote control-plane state is authoritative.

Local state exists to make the CLI responsive and useful.

That is the underlying rule behind all of the flows in this repository.

## Direct Command Flow

Some commands are straightforward. They parse input, resolve configuration and
auth, call the right service or client boundary, and render the result.

Those commands do not need a long-running local runtime to be correct. That is
why command handlers should stay small and read like short pipelines.

The entrypoint for that surface is
[`internal/interfaces/cli/execute.go`](/Users/ben/Code/usetero/cli/internal/interfaces/cli/execute.go).

## Bootstrap Flow

Before account-scoped local runtime exists, the app cannot assume there is local
synced state to work with. That is why onboarding and related startup behavior
use a different path.

The basic shape is:

1. read preferences and environment state,
2. fetch the remote state needed to understand where the user is,
3. project the current bootstrap state,
4. derive the next step from that projection,
5. perform mutations through the correct service boundaries,
6. re-read and re-project.

The important idea is that progression should come from projected state, not
from incidental widget state in the TUI.

## Steady-State Local Runtime Flow

Once bootstrap is complete, the app can switch to the account-scoped local
runtime.

That usually means:

1. resolve account scope,
2. open the scoped SQLite database,
3. start PowerSync,
4. start the uploader,
5. read local state through local services and read models,
6. keep local and remote state converging over time.

This is the path that makes the TUI feel local and fast.

## Why Some Domains Have Local And Remote Services

In some domains, the CLI needs both:

- a remote path against the control plane,
- a local path against synced SQLite state.

That is why you will see both local and remote service implementations in the
same conceptual area. The duplication is not the point. The point is that the
product has two legitimate ways to access the same kind of data depending on the
phase and the user experience you need.

## Why Queries Live Close To Callers

The repository does not use a large ORM to centralize every query into one
generic abstraction.

Instead, local queries usually live close to the package that owns the read.
That may be a domain package. It may be a read model. This works well here
because many local reads exist for one specific screen or one specific
interaction need.

Keeping those queries close to the caller makes the ownership clearer and avoids
overly generic data APIs.

## Common Failure Modes

The most common data-flow mistakes in this repo are:

- treating local data like authority,
- writing bootstrap code that assumes runtime already exists,
- letting the TUI infer truth from local widget state,
- bypassing the intended service or runtime boundary for a write,
- hiding account scope in mutable global state.

When one of those happens, the fix is usually to restore the intended flow, not
to patch around the symptom.
