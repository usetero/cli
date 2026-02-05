# Observability

Structured logging with hierarchical component scopes.

## The Mental Model

**Scopes are for components, not operations.** Unlike OpenTelemetry where spans trace individual requests or function calls, our scopes trace the component hierarchy. A scope is created once in a constructor and lives for the lifetime of that component.

Components receive a scope from their parent and create a child scope in their constructor. This builds a hierarchy that shows where logs come from:

```
app/chat/messagelist/turn/query: query completed rows=5
```

That tells you instantly: this log came from a query tool, inside a turn, inside the message list, inside chat, inside the app.

## Scopes

A scope wraps a logger with path tracking. It has the same methods as a logger (`Info`, `Error`, `Debug`, `Warn`) plus `Child()` for creating nested scopes.

```go
type Scope struct {
    logger Logger
    path   string
}
```

### Creating the Root Scope

Only one place creates a root scope - the entry point (`execute.go`):

```go
scope := log.RootScope(log.New(level))
rootCmd := NewRootCmd(scope, version)
```

### Creating Child Scopes

Every component receives a scope and creates its own child in the constructor:

```go
func New(scope log.Scope, ...) *Model {
    scope = scope.Child("chat")  // Creates "parent/chat"
    
    return &Model{
        scope:       scope,
        messageList: messagelist.New(scope, ...),  // Pass scope, child creates own
    }
}
```

The child component does the same:

```go
func New(scope log.Scope, ...) *Model {
    scope = scope.Child("messagelist")  // Creates "parent/chat/messagelist"
    
    return &Model{
        scope: scope,
    }
}
```

### Using Scopes

Log with the scope just like a logger:

```go
m.scope.Info("query completed", "rows", len(rows))
m.scope.Error("failed to execute", "error", err)
m.scope.Debug("parsing input", "raw", input)
```

Output includes the scope path automatically:

```
level=INFO scope=app/chat/messagelist/turn/query msg="query completed" rows=5
```

### When to Use What

**`Child()` - New component or subsystem**

Use when creating a distinct component that will have its own identity in logs:

```go
// In constructor - component gets its own scope
func New(scope log.Scope) *Model {
    scope = scope.Child("chat")
    return &Model{scope: scope}
}

// Passing to a child component
m.messageList = messagelist.New(m.scope, ...)  // Child creates its own
```

**`With()` - Persistent context for this scope**

Use when adding context that should appear in ALL subsequent logs from this scope. Reassign the scope:

```go
// Context that persists for the lifetime of this operation
m.scope = m.scope.With("conversation_id", convID)
m.scope.Info("started")      // has conversation_id
m.scope.Info("completed")    // has conversation_id
```

**Inline fields - One-time structured data**

Use for data specific to a single log call. Don't reassign:

```go
// Data for just this log entry
m.scope.Info("query completed", "rows", len(rows), "duration_ms", elapsed)
m.scope.Error("failed", "error", err)
```

### Decision Guide

| Situation | Use |
|-----------|-----|
| New component in constructor | `scope.Child("name")` |
| Passing scope to child component | Pass scope, let child call `Child()` |
| Context for all future logs (IDs, names) | `scope = scope.With(...)` |
| Data for one log entry | Inline: `scope.Info("msg", "key", val)` |

## Rules

1. **One root scope** - Created in `execute.go` from a logger
2. **Components create their own child** - Don't pass a pre-named scope, let the component name itself
3. **Scope is concrete** - Not an interface, just pass `log.Scope`
4. **Child in constructor only** - Call `scope.Child("name")` at the start of the constructor, nowhere else
5. **Methods use the component's scope** - No new scopes in methods, just log with `m.scope.Info(...)`

## Not OpenTelemetry

Our scopes are simpler than OTel spans:

| OTel Spans | Our Scopes |
|------------|------------|
| Created per request/operation | Created per component |
| Short-lived (request duration) | Long-lived (component lifetime) |
| Trace distributed systems | Trace component hierarchy |
| Have start/end times | No timing, just path |

We're solving a different problem: "which component logged this?" not "how long did this request take?"

## Example

```go
// execute.go - creates root
scope := log.RootScope(log.New(level))
rootCmd := NewRootCmd(scope, version)

// root.go - passes to app
app.New(ctx, ..., scope)

// app.go - creates child, passes to children
func New(..., scope log.Scope) *Model {
    scope = scope.Child("app")
    return &Model{
        scope:      scope,
        onboarding: onboarding.New(..., scope),
        chat:       chat.New(..., scope),
    }
}

// chat.go - creates child, passes to children
func New(..., scope log.Scope) *Model {
    scope = scope.Child("chat")
    return &Model{
        scope:       scope,
        messageList: messagelist.New(..., scope),
    }
}

// Logs show the full path:
// scope=app msg="started"
// scope=app/chat msg="initialized"
// scope=app/chat/messagelist msg="loading messages"
```

## Testing

Use `logtest.NewScope(t)` to create a test scope:

```go
func TestQuery(t *testing.T) {
    scope := logtest.NewScope(t)
    model := query.New(scope, ...)
    // Logs appear in test output only on failure or with -v
}
```
