# Domains

Domain docs answer a different question than architecture docs.

Architecture docs explain where responsibilities belong. Domain docs explain
what behavior is expensive to get wrong once it is in the right place.

Read this section when you are:

- changing a user-visible workflow,
- modifying progression or lifecycle semantics,
- tightening or relaxing product invariants,
- trying to understand which failures matter most in a specific flow.

## What belongs in this section

A domain page is worth keeping only when a product area has all of these:

- a clear behavioral model,
- non-trivial failure modes,
- invariants that are easy to regress with local edits,
- a small set of code entry points worth teaching explicitly.

This section should not mirror the entire package tree.

## Current domain coverage

- [onboarding-bootstrap.md](onboarding-bootstrap.md): deterministic bootstrap
  progression before account-scoped runtime exists.
- [chat.md](chat.md): conversation/message ownership, streaming runtime
  lifecycle, and tool execution boundaries.
- [statusbar.md](statusbar.md): shell-level session status presentation and the
  boundary between runtime truth and compact terminal chrome.

## How to use these docs

Use the domain page first, then confirm the underlying boundary in:

- [`../architecture/system-overview.md`](../architecture/system-overview.md)
- [`../architecture/data-flow.md`](../architecture/data-flow.md)
- [`../architecture/runtime-architecture.md`](../architecture/runtime-architecture.md)

That sequence keeps behavioral edits from quietly turning into boundary leaks.
