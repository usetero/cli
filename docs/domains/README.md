# Domains

Domain docs answer a different question than architecture docs.
Architecture tells you where code should live. Domain docs tell you who owns
behavior and which invariants are expensive to violate.

Read this section when you are changing product behavior, not just moving code.

## Current domain coverage

[onboarding-bootstrap.md](onboarding-bootstrap.md) explains bootstrap flow
ownership and deterministic gate contracts.

[chat.md](chat.md) explains the split between stream protocol correctness and
UI orchestration semantics.

As new high-risk product areas appear, they should get a domain page only when
there is a real invariant set worth protecting. This section is not meant to
mirror the entire package tree.

## How to use these docs in practice

Before a behavioral change:

1. read the relevant domain page,
2. confirm the corresponding architecture boundary,
3. identify the invariant that must remain true after your change.

This sequence prevents “works locally” fixes that regress user-visible runtime
semantics in adjacent areas.
