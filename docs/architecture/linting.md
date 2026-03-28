# Linting And Guardrails

This repository should treat linting as executable architecture doctrine.

The goal is not to encode every engineering preference as a failing check. The
goal is to mechanically protect the small set of rules that keep the codebase
boring, obvious, and structurally honest.

## The standard

This repo has stronger boundaries than a typical Go CLI:

- interfaces present and compose
- runtime coordinates long-lived workflows
- domains own product-shaped operations and invariants
- infrastructure implements concrete capabilities

Because those boundaries matter, linting should enforce architecture first,
repo-specific runtime rules second, and generic style last.

## What belongs in lint

Rules belong in lint when they are:

- stable
- recurring
- mechanically legible
- expensive to keep re-teaching in review

That includes:

- layer boundaries
- TUI event-loop discipline
- parent/child Bubble Tea ownership
- dependency naming conventions
- path and docs drift

## What does not belong in lint

Rules should stay in docs or review when they require deeper judgment about
ownership or product meaning.

Examples:

- whether a new runtime exists at the right boundary
- whether a flow belongs in onboarding or steady-state runtime
- whether a new domain type is the right product abstraction

Lint should protect doctrine, not replace engineering judgment.

## The target end state

The final shape should be:

- one repo-owned lint entrypoint
- a small analyzer bundle under `scripts/lint/analyzers`
- narrow shell checks only where AST analysis is not worth the cost
- clear failure messages
- explicit baselines for current drift instead of weakening the target rules

The important point is that the rules stay strict even when the current repo is
still moving toward them.

## Current doctrine to protect

The high-value rules for this repo are:

- `domains` must not import concrete `infrastructure`
- `runtime` must not import `interfaces`
- `infrastructure` must not import `runtime` or `interfaces`
- Bubble Tea `tea.Cmd` closures must emit messages only
- blocking or external work must not happen directly in `Update` or `View`
- flow parents own routing; leaf models own only local interaction state

Those are the rules that keep the architecture obvious under change.

## Baselines

Baselines are acceptable only as a ratchet.

If the codebase already violates a target rule, record the violating files
explicitly and keep the analyzer strict. That makes drift visible, prevents new
violations, and keeps the intended end state clear.

The baseline should shrink over time. It should never become a hidden policy
engine for exceptions.
