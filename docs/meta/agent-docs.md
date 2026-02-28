# Agent Docs Standards

`AGENTS.md` is not a human architecture guide. It is an operational constraint
file for coding agents.

Use it to prevent mistakes, not to explain the whole system.

## What goes where

Put correctness-critical directives and workflow constraints in `AGENTS.md`.
Put conceptual architecture and behavioral explanation in `docs/`.

If a fact is only in `docs/` but an agent needs it to avoid wrong code, that
fact belongs in `AGENTS.md` too.

This division keeps `AGENTS.md` short and enforceable while allowing the main
docs tree to carry deeper explanation for humans.

## Writing bar for agent instructions

A strong instruction is:

1. precise enough to execute,
2. scoped to the directory/runtime it governs,
3. short enough to stay high-signal,
4. directly tied to known failure modes.

Avoid broad prose, repeated guidance, and information that is obvious from
local code symbols alone.
Favor concrete directives tied to actual failure modes seen in this repo.
