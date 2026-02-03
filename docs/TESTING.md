# Testing

Test behavior, not implementation. A good test breaks when something stops working, not when you refactor.

## Three Types of Tests

**Unit tests** run fast with no external dependencies. They mock interfaces and test logic in isolation.

```bash
task test
```

**Integration tests** hit real services. They require credentials and catch bugs that mocks hide.

```bash
task test:integration
```

**Correctness tests** verify external services behave as documented. They test assumptions, not our code.

```bash
task test:correctness
```

## Writing Tests

### Test Behavior

Bad: test that a field gets set.
Good: test that `HasError()` returns true after an error.

The first breaks when you rename the field. The second only breaks when behavior changes.

### Use Subtests

One test function per method. Use `t.Run()` for scenarios:

```go
func TestSync_Connect(t *testing.T) {
    t.Run("returns auth error on 401", func(t *testing.T) {
        // ...
    })
    
    t.Run("retries on 503", func(t *testing.T) {
        // ...
    })
}
```

### Mock with Structs

Don't use mocking frameworks. Write simple structs:

```go
type mockClient struct {
    doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
    return m.doFunc(req)
}
```

### Choose Internal or External Package

External package (`package foo_test`) tests the public API. Prefer this when it gives you confidence—it survives refactoring better.

Internal package (`package foo`) can access unexported functions. Use this when you need to test internals that matter.

Pick whichever gives you more confidence. It's not a hard rule.

### Run Tests in Parallel

```go
func TestSync_Connect(t *testing.T) {
    t.Parallel()
    
    t.Run("returns auth error on 401", func(t *testing.T) {
        t.Parallel()
        // ...
    })
}
```

Exception: tests using `t.Setenv()` can't run in parallel.

### Use Test Helpers

Use the `*test` packages for common test doubles:

```go
import (
    "github.com/usetero/cli/internal/log/logtest"
    "github.com/usetero/cli/internal/chat/chattest"
    "github.com/usetero/cli/internal/sqlite/sqlitetest"
)

func TestFoo(t *testing.T) {
    logger := logtest.New(t)  // logs appear on test failure
    client := &chattest.MockClient{...}
    db := sqlitetest.OpenBareDB(t)
}
```

`logtest.New(t)` is preferred over discarding logs—you'll want them when debugging failures.

## File Naming

| Type | File | Build Tag | Function Prefix |
|------|------|-----------|-----------------|
| Unit | `foo_test.go` | none | `Test` |
| Integration | `foo_integration_test.go` | `//go:build integration` | `TestIntegration_` |
| Correctness | `foo_correctness_test.go` | `//go:build correctness` | `TestCorrectness_` |

## What to Test

Test:
- State transitions
- Error handling
- Coordination between components
- Edge cases that could break

Skip:
- One-line delegators
- Simple getters
- Framework behavior
- Code you're about to delete

## What Makes Tests Good

- Fast (mock external deps)
- Deterministic (no randomness, no timing)
- Clear failures (you know what broke)
- Resilient (survive refactoring)

The goal is confidence. When tests pass, you ship.
