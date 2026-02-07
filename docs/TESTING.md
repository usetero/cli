# Testing

Test behavior, not implementation. A good test breaks when something stops working, not when you refactor.

## Three Types of Tests

**Unit tests** run fast with no external services. Use real local deps (SQLite, theme), mock remote ones (APIs).

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

### Real vs Mock

Use real things when they're cheap. Mock only what's expensive or external.

| Dependency | Strategy | Why |
|------------|----------|-----|
| SQLite | Real — `sqlitetest.OpenBareDB(t)` or `dbtest.OpenTestDB(t)` | Local, fast, tmpdir-based, auto-cleanup |
| Theme | Real — `styles.NewTheme(true)` | Just a struct with colors |
| Logger | Real — `logtest.NewScope(t)` | Writes to `testing.T`, shows on failure |
| Chat API | Mock — `chattest.MockClient{}` | Hits remote server |
| Tool registry | `nil` or partial — `&chattools.Registry{Query: ...}` | Only construct what you're testing |

The rule: if you can construct it in a test without network, credentials, or slow setup, use the real thing. Mocks hide bugs.

### Mock with Structs

When you do need mocks, don't use mocking frameworks. Write simple structs:

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

## Testing `View()`

`View()` is a public method that returns a string. Test it like any other method — given state and inputs, assert the output behaves correctly.

The one TUI-specific detail: output contains ANSI escape codes, so use `ansi.StringWidth` instead of `len()` to measure visible width. The `teatest` package provides assertion helpers for this.

### Width Contract

Every component must render within its assigned width. This is the core behavioral invariant for `View()`:

```go
func TestQuery_View(t *testing.T) {
    t.Run("respects width", func(t *testing.T) {
        m := New(theme, scope)
        m.SetWidth(80)
        // populate with data...

        teatest.AssertMaxWidth(t, 80, m.View())
    })
}
```

Test at multiple widths to catch edge cases:

```go
for _, width := range []int{40, 80, 120, 200} {
    t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
        // ...
        teatest.AssertMaxWidth(t, width, m.View())
    })
}
```

### Test the Real Rendering Chain

Parent components wrap child `View()` output with padding, borders, etc. Test with real models to catch width accounting bugs between layers:

```go
// BAD: simulating the parent's rendering with manual lipgloss calls
body := lipgloss.NewStyle().PaddingLeft(2).Render(childView)

// GOOD: using real models — catches disagreements between parent and child
parent := assistant.New(theme, "id", width, nil, scope)
parent.AddBlock(realToolModel)
teatest.AssertMaxWidth(t, width, parent.View())
```

### Test Construction-Time Rendering

Components may render before `SetWidth` is called (e.g. a query completes during streaming before any resize event). Test both paths:

```go
// Path 1: construction-time width only (no SetWidth call)
m := New(theme, "id", width, nil, scope)
m.AddBlock(tool)
teatest.AssertMaxWidth(t, width, m.View())

// Path 2: after explicit SetWidth
m.SetWidth(newWidth)
teatest.AssertMaxWidth(t, newWidth, m.View())
```

### ANSI Integrity

Components that render styled text can break ANSI escape sequences when slicing strings by byte offset instead of visible position. Use `AssertNoRawEscapes` to catch this:

```go
func TestView(t *testing.T) {
    t.Run("no raw escape sequences", func(t *testing.T) {
        m := New(theme)
        m.SetWidth(50)

        teatest.AssertNoRawEscapes(t, m.View())
    })
}
```

This catches bugs like `view[:bytePos]` slicing through a `\x1b[38;2;110;231;183m` color code, leaving raw `38;2;110;231;183m` visible to the user.

### Test Helpers

| Package | Import | Use |
|---------|--------|-----|
| `teatest` | `github.com/usetero/cli/internal/tea/teatest` | `AssertMaxWidth()` — no line exceeds width; `AssertExactWidth()` — widest line equals width; `AssertNoRawEscapes()` — no broken ANSI sequences |
| `logtest` | `github.com/usetero/cli/internal/log/logtest` | `NewScope(t)` for test loggers |
| `styles` | `github.com/usetero/cli/internal/styles` | `NewTheme(true)` for dark theme in tests |

### Reference Tests

- `assistant_test.go` — full rendering chain: assistant → tool → query
- `query_test.go` — single component width testing at multiple widths
- `palette_test.go` — ANSI integrity + width at multiple sizes
- `input_test.go` — cursor marker insertion doesn't break escapes

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
