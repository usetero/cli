# What This Repo Is

This repository is the CLI for Tero.

That includes the interactive terminal UI, the direct command surface, and the
MCP server. More importantly, it means this repository is the presentation layer
for the Tero control plane. It is responsible for turning control-plane state
and local synced state into a terminal experience that feels clear, responsive,
and useful.

That description is important because it explains what belongs here and what
does not.

## What The CLI Owns

The CLI owns the user-facing experience in the terminal.

That includes:

- command parsing and command output,
- TUI composition and interaction,
- local presentation state,
- account-scoped runtime startup and shutdown,
- local read paths that make the interface fast.

This is real product code. It is not just a thin wrapper script over an API.

## What The CLI Does Not Own

The CLI does not own the central business truth of Tero.

The control plane is the source of truth for the product. The CLI presents that
truth, works with it, and in some cases keeps a synced local copy of parts of it
to make the terminal experience better. But it should not quietly become a
second authority.

That is the boundary to keep in mind throughout the repository.

## Why The Repo Has More Structure Than A Typical CLI

If this were only a set of direct commands, the repository could be much
smaller.

But the product needs more than one-off requests. The TUI needs to feel local.
Some screens need to explore synced data quickly. Bootstrap flows need to work
before local runtime exists. Long-lived account-scoped runtime has to manage
SQLite, PowerSync, and upload behavior over time.

That is why the repository has distinct runtime, read-model, domain, and
infrastructure layers instead of only command handlers.

## How To Think About It

The simplest useful way to think about this repo is:

the control plane owns product truth, and the CLI turns that truth into a good
terminal experience.

Sometimes that means direct remote calls.
Sometimes that means a local synced database.
Sometimes that means a read model that exists only to keep the TUI clean.

The common theme is the same: the CLI should make the product easier to use
without redefining it.

## Where The Program Starts

The main entrypoint is
[`cmd/tero/main.go`](/Users/ben/Code/usetero/cli/cmd/tero/main.go), which hands
off to the CLI interface under
[`internal/interfaces/cli`](/Users/ben/Code/usetero/cli/internal/interfaces/cli).

From there, the program selects one of the user-facing surfaces:

- the TUI,
- direct commands,
- MCP.

Those surfaces are different entrypoints into the same system. They should feel
appropriate for the medium, but they should still reflect the same product
model.

## Why This Matters

Most bad changes in this repo come from getting the ownership model wrong.

If interface code starts deciding business truth, or local state starts being
treated like authority, or infrastructure starts taking over workflow semantics,
the repository gets harder to reason about very quickly.

So before you read anything else, keep the core idea in mind:

the CLI is the presentation runtime for the control plane, not a separate
product brain.
