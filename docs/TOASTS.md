# Toasts

Toasts are transient notifications that appear at the bottom of the screen to communicate status to users.

## When to Use Toasts

**Use toasts** when you want the user to see feedback:
- Operation failed and user should know
- Operation succeeded and user should know
- Warning the user about something

**Use logging** when you don't need to bother the user:
- Internal errors that don't affect UX
- Debug information
- Errors that are handled silently

## Message Types

Defined in `internal/app/msgs/msgs.go`:

| Type | Use Case | Sticky Option |
|------|----------|---------------|
| `Error` | Operation failed | Yes |
| `Warning` | Something concerning | Yes |
| `Success` | Operation succeeded | No |
| `Info` | Neutral information | No |

## Usage

Import the package and use the `Cmd` constructors:

```go
import appmsg "github.com/usetero/cli/internal/app/msgs"

// In Update(), return a command:
if err != nil {
    return m, appmsg.ErrorCmd("Failed to save", err, false)
}
return m, appmsg.SuccessCmd("Saved!")
```

### Constructors

```go
appmsg.ErrorCmd(message string, err error, sticky bool) tea.Cmd
appmsg.WarningCmd(message string, sticky bool) tea.Cmd
appmsg.SuccessCmd(message string) tea.Cmd
appmsg.InfoCmd(message string) tea.Cmd
```

### Sticky vs Non-Sticky

- **Non-sticky** (default): Auto-dismisses after 5 seconds
- **Sticky**: Stays until user action or another toast replaces it

Use sticky for errors that block progress. Use non-sticky for transient feedback.

## How It Works

1. Component returns a `Cmd` constructor (e.g., `appmsg.ErrorCmd(...)`)
2. Bubbletea executes the `Cmd`, producing a message (e.g., `appmsg.Error{...}`)
3. Message bubbles up through the component tree
4. Toast component (`internal/app/toast`) receives all messages
5. Toast matches on message types and renders accordingly

The toast component is forwarded all messages in `app.go`:

```go
// Forward to current state and toast
var cmds []tea.Cmd

if cmd := m.toast.Update(msg); cmd != nil {
    cmds = append(cmds, cmd)
}
// ... forward to other components
```

## Files

- `internal/app/msgs/msgs.go` - Message types and constructors
- `internal/app/toast/toast.go` - Toast component (state + rendering)
