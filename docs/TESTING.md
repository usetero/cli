# Testing

## Introduction

Good tests make you confident to ship. Bad tests make you scared to change anything. The difference isn't coverage percentages or test counts—it's whether your tests catch real bugs without breaking every time you refactor.

This matters because the CLI is going to grow. More pages, more components, more ways for things to go wrong. Without good tests, every change becomes risky. With good tests, you can refactor with confidence, knowing that if you break something that matters, the tests will tell you.

This document explains how we think about testing in the CLI. Not rules to memorize, but principles to internalize. When you understand why we test certain things and ignore others, you'll write tests that actually help—tests that catch bugs, survive refactoring, and make the codebase easier to work with.

## What Makes a Test Good

A good test tells you when behavior changes. A bad test tells you when implementation changes. The difference is subtle but crucial.

Imagine you have a step that validates user input. You could test that it has a field called `validationErr` that gets set when validation fails. That's testing implementation. Or you could test that after receiving a validation failure message, calling `HasError()` returns true and `Error()` returns the error message. That's testing behavior.

The implementation test breaks if you rename the field, change how errors are stored, or refactor the validation logic—even if the behavior stays exactly the same. The behavior test only breaks if validation actually stops working. One forces you to update tests during harmless refactoring. The other lets you refactor freely and only fails when you've actually broken something.

This distinction runs through everything we do. Test what the code *does*, not how it does it. Test the contract it fulfills, not the internal machinery. Test what users experience, not what developers implement.

## Testing a Presentation Layer

The CLI doesn't implement business logic. It doesn't analyze telemetry, calculate waste, or make decisions about data quality. All of that lives in the control plane. The CLI coordinates authentication, communicates via GraphQL, and renders interactive interfaces.

This shapes what we test. We're not testing algorithms or business rules. We're testing coordination—does authentication flow work correctly? Do errors propagate from components up to the footer? Does state transition properly when messages arrive? Do views reflect the current state?

Think about what could actually go wrong in a presentation layer:
- A user presses Enter but nothing happens (message not handled)
- An error occurs but doesn't display (state not propagated)
- A loading indicator never stops (state not cleared)
- Retry doesn't work (error state not reset)
- View shows wrong content for current state (view logic broken)

These are the bugs that matter. These are what we test.

What we don't test: that Bubbletea delivers messages correctly, that lipgloss renders colors right, that the GraphQL client serializes JSON properly. Those are dependencies, and we trust them. Testing them would be testing someone else's code, not ours.

## The Shape of TUI Tests

The TUI is built on the Elm Architecture—model, update, view. This makes it naturally testable. Every component is a pure function: given a model and a message, return the new model and any commands. No hidden state, no side effects in the update function itself, just data transformation.

This means you can test without a terminal, without rendering, without any special test harness. Create a model, send it messages by calling `Update()`, check the results. It's just functions.

Here's the pattern you'll see everywhere:

First, create the model with test dependencies. Real code takes interfaces—`AuthService`, `AccountCreator`, `Logger`. Tests pass in simple mocks that do exactly what the test needs.

Second, send messages. Call `Update()` with the message you're testing. If that returns a command, execute it to get the next message and send that too. You're simulating the message flow that would happen in the real app.

Third, check state. Call the public methods—`HasError()`, `IsBusy()`, `IsComplete()`. These tell you what state the model is in. If you sent an error message, `HasError()` should return true. If you started an async operation, `IsBusy()` should return true.

Fourth, optionally check the view. Call `View()` and look for key content. Don't check exact strings or styling—those change. Check that the right information appears. If there's an error, does something in the view reflect that? If it's loading, does the view indicate that?

The pattern repeats at every level. Testing a step? Create it, send messages, check state. Testing a mode that orchestrates steps? Same pattern, just one level up. Testing the root TUI? Same pattern, just testing coordination across modes.

## State and Propagation

The hardest bugs to catch are propagation bugs. A step sets error state correctly, but the mode doesn't pass it to the layout, so the footer never shows it. Or the mode reads error state *before* updating the step, so it's always one message behind.

These bugs are subtle because each piece works in isolation. The step correctly sets its error. The mode correctly passes errors to the layout. But the *timing* is wrong, and the error doesn't appear when it should.

This is why we test propagation explicitly. Create a mode with a step that will error. Send the error message. Check that the mode's `HasError()` returns true. Check that calling `View()` includes the error content. You're testing the full path from step to screen.

These tests are more complex than unit tests because they involve multiple components. But they catch real bugs that unit tests miss—bugs where the pieces work individually but don't work together.

## Mocks and Dependencies

The app layer uses interfaces everywhere—`AuthService`, `PreferencesService`, `AccountCreator`, `APIClient`. This makes testing straightforward. Real code depends on interfaces, tests provide implementations that do exactly what the test needs.

Don't reach for mocking frameworks. They add complexity and obscure what's happening. Instead, write simple structs with function fields:

```go
type mockAccountCreator struct {
    createAccountFunc func(ctx context.Context, accountID, name, site, apiKey, appKey string) (*api.DatadogAccount, error)
}

func (m *mockAccountCreator) CreateAccount(ctx context.Context, accountID, name, site, apiKey, appKey string) (*api.DatadogAccount, error) {
    if m.createAccountFunc != nil {
        return m.createAccountFunc(ctx, accountID, name, site, apiKey, appKey)
    }
    return nil, errors.New("mock not configured")
}
```

Now in tests you can see exactly what behavior you're providing:

```go
creator := &mockAccountCreator{
    createAccountFunc: func(ctx context.Context, accountID, name, site, apiKey, appKey string) (*api.DatadogAccount, error) {
        return nil, errors.New("invalid application key")
    },
}
```

No magic, no surprise behavior, just explicit configuration. When you read the test, you know exactly what the mock does because you can see the function right there.

## Writing Tests That Last

Tests should survive refactoring. When you move code around, rename things, or change internal structure without changing behavior, tests should keep passing. If they don't, they're testing implementation details that don't matter.

This means you need to think about what's fundamental and what's incidental. Fundamental: a step in error state should report `HasError() == true`. Incidental: the step stores errors in a field called `err`. Test the fundamental thing, not the incidental thing.

The same applies to views. Fundamental: the view shows different content when in error state vs normal state. Incidental: the error is red text in a specific position. Test that the content changes, not the exact rendering.

This takes judgment. Sometimes you need to test something specific. But default to testing behavior over implementation, contracts over internals, what changes over how it changes.

## When Tests Matter Most

Write tests when fixing bugs. When you encounter a bug, first write a test that fails because of the bug. Then fix the bug. Now you have a test that prevents regression. This is the highest-value testing you can do—it directly prevents a bug you know happened from happening again.

Write tests for error paths. Error handling is where bugs hide. The happy path usually works. It's the network failure, the invalid input, the unexpected response that breaks. Test that errors are caught, state is set correctly, users can retry, and the UI degrades gracefully.

Write tests when refactoring complex code. If you're about to significantly change how something works, write tests first. They give you confidence that your refactoring didn't change behavior. This is especially valuable in a presentation layer where bugs are often visual—hard to spot by inspection but easy to catch with tests.

Don't write tests for code you're about to delete, simple getters that just return field values, or glue code that just wires things together. Test business logic, coordination logic, error handling, and state transitions. Skip the trivial stuff.

## Test Structure and Naming

We follow consistent patterns for organizing and naming tests. These aren't arbitrary rules—they make tests easier to find, easier to read, and easier to maintain.

### One Test Function Per Function

When testing a type's method, create one test function that matches the method name. Inside that function, use `t.Run()` to test different scenarios:

```go
func TestAppKeyStep_Update(t *testing.T) {
    t.Run("sets error state when account creation fails", func(t *testing.T) {
        // Test error scenario
    })
    
    t.Run("clears error state on successful creation", func(t *testing.T) {
        // Test success scenario
    })
    
    t.Run("ignores enter key while already creating", func(t *testing.T) {
        // Test debouncing
    })
}
```

This groups related tests together. When `Update()` breaks, you know exactly which test file and which function to look at. The `t.Run()` names describe the specific behaviors being tested.

### Table-Driven Tests for Input Variations

Use table-driven tests when you're testing the same logic with different inputs. The table should contain just inputs and expected outputs—not complex logic that drives different testing behavior:

```go
func TestValidateAccountName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid name", "production-account", false},
        {"empty name", "", true},
        {"too long", strings.Repeat("a", 256), true},
        {"invalid chars", "prod@account!", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateAccountName(tt.input)
            if tt.wantErr && err == nil {
                t.Error("expected error, got nil")
            }
            if !tt.wantErr && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

This works well when you're testing pure functions with clear inputs and outputs. Don't use tables for testing complex state machines or multi-step interactions—those need explicit test logic.

### Test Package Organization

Test helpers and mocks live in a sibling package named `{package}test` (one word). If you have a package `account`, create `accounttest` for its test utilities:

```
internal/
  account/
    account.go
    account_test.go
  accounttest/
    mocks.go
```

This keeps mocks reusable across packages without creating import cycles. The app layer follows this pattern—look at how services use interfaces and `apptest` provides mock implementations.

When a function accepts an interface, write tests that use mock implementations. This makes the function testable without real dependencies. If you find yourself wanting to mock something that's not an interface, consider whether that dependency should be extracted behind an interface.

### Naming Conventions

Test function names follow the pattern: `Test{Type}_{Method}` or `Test{Function}`. The name identifies what's being tested. Inside, `t.Run()` names describe specific behaviors in plain English:

- `TestAppKeyStep_Update` tests the Update method
- `TestAppKeyStep_HasError` tests the HasError method  
- `TestValidateAPIKey` tests the ValidateAPIKey function

The `t.Run()` names are sentences: "sets error state when account creation fails", not "error_case" or "test_2". When a test fails, you should immediately understand what behavior broke.

### What Not to Test

Some methods aren't worth testing because they're trivial or just delegate to other code you trust. When you're looking at a file and deciding what to test, ask: what could actually break here that would matter?

Trivial delegators don't need tests. If a method just calls another method and returns its result, you're testing a one-line function that can't break in any interesting way:

```go
// Don't test this
func (m *Onboarding) HasError() bool {
    return m.flow.HasError()
}

// Or this
func (m *Onboarding) Error() error {
    return m.flow.Error()
}
```

These are pass-throughs. They can't have bugs. Testing them just creates maintenance burden—if you refactor how errors work, you have to update tests that never caught bugs.

What you *do* test is the orchestration logic—the code that coordinates multiple pieces. In `Onboarding`, that's the `Update()` method that needs to propagate error state from the flow to the layout at the right time. That's where bugs hide. That's what matters.

Simple getters and setters don't need tests. Field assignments don't need tests. One-line delegators don't need tests. Focus on the logic that actually does something—the coordination, the state transitions, the error handling.

## Tests as Documentation

When someone reads your test, they should understand what the code does and why. The structure we just described helps with this—one test function per method, clear scenario names, grouped related behaviors. But it goes beyond structure.

Use clear arrange-act-assert structure. Set up your test state, perform the action, check the results. Add comments that explain *why* you're testing something, not *what* the code does—they can read the code for that.

Good tests teach. When a new contributor wants to understand how error propagation works, they should be able to read the tests and see the pattern. When someone is fixing a bug, they should be able to read tests to understand expected behavior.

This means tests need to be readable. Don't be clever. Don't compress everything into one line. Don't use obscure helpers that hide what's happening. Be explicit, be clear, and optimize for the reader.

## Running and Maintaining Tests

Tests should be fast. No network, no real authentication, no actual database. Mock external dependencies so tests run in milliseconds. This makes them useful during development—you can run them constantly as you work.

Tests should be deterministic. No randomness, no time dependencies, no order dependencies. Every run should produce the same results. Flaky tests are worse than no tests—they train you to ignore failures.

When tests fail, they should tell you clearly what broke. Use assertions that give good error messages. Better yet, structure tests so that when something fails, you can immediately see what was expected and what happened.

As the codebase grows, some tests will become obsolete. Delete them. Tests have maintenance cost. If a test no longer serves a purpose, remove it. Don't accumulate tests just to increase coverage numbers.

## The Goal

The goal is confidence. You should feel good merging code when tests pass. You should catch bugs before users do. But tests shouldn't slow you down or make refactoring painful.

Test what matters: behavior, not implementation. State transitions, not field assignments. Error handling, not happy paths. Propagation, not isolated units.

Trust your tools. Don't test the framework. Don't test dependencies. Test your code—the coordination logic, the state management, the error handling that you wrote.

Keep tests simple. Use mocks, not frameworks. Test behavior, not structure. Write tests that survive refactoring by testing contracts, not internals.

When you write a test, ask: does this catch a real bug? Does it test behavior that matters to users? Will it survive refactoring? If yes, write it. If no, skip it.

We're building a TUI, not a space shuttle. Be pragmatic. Test what matters, skip what doesn't, and ship with confidence.
