# Testing

Testing in this repository should buy confidence, not drag.

Tests are code too. They can clarify behavior and make changes safer, or they
can become brittle and make the codebase harder to evolve. The standard here is
simple: write tests that are worth keeping.

## Start With Behavior

The first question is not "how do I get coverage here?"

The first question is:

what behavior matters, and which layer should prove it?

That is the center of gravity for testing in this repo.

Good tests protect behavior that should survive refactors:

- a workflow progresses correctly,
- a service validates and maps correctly,
- a runtime publishes the right state,
- a screen routes and reacts correctly,
- an end-to-end path actually works.

Bad tests lock down incidental implementation detail and make harmless code
changes expensive.

## The Main Test Layers

There are a few real testing layers in this repo.

### Unit tests

Most tests in the repository are unit tests around one package or one focused
behavior.

These should live close to the code they prove and should target the layer that
actually owns the behavior:

- interface tests for UI or command behavior,
- runtime tests for progression and lifecycle,
- domain tests for business-shaped behavior,
- infrastructure tests for concrete adapter behavior,
- read-model tests for local shaping.

### Executable integration and end-to-end tests

[`cmd/tero`](/Users/ben/Code/usetero/cli/cmd/tero) is also where the
high-confidence binary tests live.

Those tests are different from unit tests. They prove that a user-facing binary
path actually works end to end.

The current onboarding tests show the split clearly:

- [onboarding_smoke_test.go](/Users/ben/Code/usetero/cli/cmd/tero/onboarding_smoke_test.go)
  proves a high-level binary path against fake/local dependencies,
- [onboarding_e2e_test.go](/Users/ben/Code/usetero/cli/cmd/tero/onboarding_e2e_test.go)
  proves a real end-to-end onboarding path against real services.

These tests are high value because they give confidence that the whole path
works, not just that individual packages behave in isolation.

## File And Test Naming

Keep test naming boring and predictable.

When you are writing unit tests:

- the test file should usually match the file it is testing,
- the test function should usually match the method or behavior it is proving,
- use `t.Run` to organize meaningful subcases when the same method has multiple
  variants worth exercising.

This repo already follows that pattern in many places. It is easier to scan,
easier to maintain, and easier to extend than creative naming schemes.

The goal is that a reader can go from production code to test code without
needing to guess where the relevant test lives.

## Mocks Live In Nested `*test` Packages

Reusable mocks should live in a nested test package next to the thing they
support.

Examples already in the repo include:

- [internal/domains/preferences/preferencestest](/Users/ben/Code/usetero/cli/internal/domains/preferences/preferencestest)
- [internal/domains/tenancy/tenancytest](/Users/ben/Code/usetero/cli/internal/domains/tenancy/tenancytest)
- [internal/domains/integrations/integrationstest](/Users/ben/Code/usetero/cli/internal/domains/integrations/integrationstest)
- [internal/infrastructure/sqlite/sqlitetest](/Users/ben/Code/usetero/cli/internal/infrastructure/sqlite/sqlitetest)

This pattern is important.

It makes test infrastructure first class without scattering mocks through random
test files or centralizing them in a grab bag package nobody owns clearly.

## Mock Style

Mocks in this repo should stay simple.

The preferred pattern is a functional mock with one `*Func` field per method,
plus a safe default when the function is unset.

The preferences mock is a good example:

- [internal/domains/preferences/preferencestest/mock_service.go](/Users/ben/Code/usetero/cli/internal/domains/preferences/preferencestest/mock_service.go)

That style works well because it is:

- explicit,
- easy to read in tests,
- easy to customize per case,
- easy to extend when an interface grows.

Avoid mocking frameworks. They add indirection without helping much in a codebase
like this.

## Use Real Local Databases

Do not mock the local database.

If code is working against SQLite or another local storage boundary in-process,
it is usually easier and more trustworthy to use the real thing in tests.

That applies to local services, infrastructure code, and other logic where the
database behavior is part of what you actually care about.

Mock network boundaries. Do not mock SQLite just because it is technically
possible.

The local database is fast enough that using the real thing is usually the
simpler and better option.

## What Different Layers Should Prove

### Interface tests

Interface tests should prove interaction behavior.

For the TUI, that usually means:

- route ownership,
- message forwarding,
- busy and error transitions,
- typed-message updates,
- rendering contracts that matter for navigation.

Do not over-test cosmetic rendering details that would make refactors painful
without increasing confidence.

### Runtime tests

Runtime tests should prove progression, lifecycle, and state ownership.

The onboarding runtime is the clearest example. Tests there should prove that
the next step comes from projected state, not from incidental UI behavior.

The account runtime should prove lifecycle behavior, readiness, and event/status
publication boundaries.

### Domain tests

Domain tests should prove business-shaped behavior.

That includes:

- typed input validation,
- service behavior,
- local-versus-remote behavior where the domain owns the distinction,
- mapping and normalization that matter for the product.

### Infrastructure tests

Infrastructure tests should prove the concrete adapter does what it says it
does.

That includes things like:

- request construction,
- response mapping,
- auth behavior,
- classification of remote errors,
- SQLite behavior and transaction handling,
- PowerSync syncer/uploader invariants.

### Read-model tests

Read-model tests should prove that local data is shaped correctly for the
surface that consumes it.

They are not proving domain truth. They are proving the presentation contract
over local state.

## End-To-End Tests Should Be High Confidence

End-to-end tests are expensive enough that they should prove something valuable.

The point of these tests is high confidence that a path really works from one end
to the other. That means they should use real services when the whole point is
to validate the real path.

In this repo there are two main ways to do that:

- run against a local control-plane stack,
- run against the shared remote dev environment.

For stable non-destructive paths, using the existing demo or dev environment is
often the easiest way to get real coverage without spinning everything up
locally.

## Integration Tests Against Real Services

Real-service integration tests should be written with the expectation that the
environment is alive and the data is not perfectly static.

That means the assertions should be:

- meaningful,
- high-signal,
- not brittle about incidental data changes.

The test should prove the path is working, not that every byte of the remote
environment is frozen forever.

## What Not To Test

Do not spend the test budget on low-value assertions that mostly prove obvious
programmer mistakes fail fast.

Examples:

- trivial constructor panics for clearly required dependencies,
- tests that only restate implementation detail,
- overly brittle rendering snapshots,
- mocks of local storage boundaries where the real thing is simpler.

If a test will make refactors harder without protecting important behavior, it
is probably not worth keeping.

## A Good Testing Standard

A good test in this repo should do at least one of these well:

- protect a user-visible workflow,
- protect a runtime invariant,
- protect a domain contract,
- protect a local read shape,
- prove a real end-to-end path.

That is the bar.

## Onboarding Test Matrix

Onboarding is worth treating as a scenario matrix instead of a grab bag of
tests.

The important user-visible states are:

- first-time user with no orgs,
- returning user with one org, account, and workspace,
- stale organization, account, or workspace preferences,
- existing Datadog account that should skip setup,
- incomplete Datadog setup that should continue from the right step,
- expired local auth that should route back to sign-in,
- real demo-account reads against the shared environment.

Those scenarios should be split by ownership layer:

- [`internal/runtime/onboarding`](/Users/ben/Code/usetero/cli/internal/runtime/onboarding)
  proves progression and state projection,
- [`internal/interfaces/tui/screens/onboarding`](/Users/ben/Code/usetero/cli/internal/interfaces/tui/screens/onboarding)
  proves route ownership, busy and error transitions, and typed-message
  handling,
- [`internal/infrastructure/controlplane/api`](/Users/ben/Code/usetero/cli/internal/infrastructure/controlplane/api)
  and remote domain adapters prove request construction, scoping, and response
  mapping,
- [`cmd/tero`](/Users/ben/Code/usetero/cli/cmd/tero) proves a small number of
  executable onboarding paths with fake or real services.

The useful rule is:

test the onboarding decision in the runtime, test the rendering reaction in the
TUI, test the network contract in the adapter, and only keep a few binary paths
that prove the whole thing actually works.

That keeps the suite predictable:

- runtime tests are the scenario table,
- TUI tests stay narrow,
- adapter tests catch scoping and mapping bugs,
- smoke and live integration tests protect the highest-value paths.
