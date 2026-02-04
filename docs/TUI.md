# TUI

The terminal UI, built on [Bubbletea](https://github.com/charmbracelet/bubbletea). This is where most complexity lives.

## Bubbletea in 30 Seconds

Bubbletea uses the Elm architecture: Model, Update, View.

```go
type Model struct {
    count int
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        if msg.String() == "+" {
            m.count++
        }
    }
    return m, nil
}

func (m Model) View() string {
    return fmt.Sprintf("Count: %d", m.count)
}
```

- **Model** holds state (value type, immutable-style)
- **Update** handles messages, returns new model + optional command
- **View** renders model to string
- **Cmd** is async work that produces a message when done

That's it. Everything else is composition.

## Model Hierarchy

The TUI is a tree of models. Each owns its children.

```
tui.Model (root)
├── onboarding.Model (auth, account selection, sync)
│   └── step.Flow (manages step transitions)
│       ├── auth/check.Model
│       ├── auth/authenticate.Model
│       ├── account/select.Model
│       ├── workspace/select.Model
│       └── sync.Model
│
└── app.Model (main app after onboarding)
    └── chat.Model (the chat page)
        ├── headerLayout / baseLayout
        ├── commandBar (input)
        ├── messages.Model (scrollable list)
        │   └── []Item
        │       ├── user.Model
        │       ├── assistant.Model
        │       └── tools.Model
        │           └── Body (query, startjourney, etc.)
        ├── sidebar
        └── thinking.Model
```

**Root model** (`internal/tui/model.go`) decides: onboarding or app?
- Not authenticated → onboarding
- Onboarding complete → transition to app

**App model** (`internal/tui/app/model.go`) is a thin router. Currently just delegates to chat.

**Chat model** (`internal/tui/app/chat/model.go`) is where the action is.

## Models All The Way Down

Every piece of UI is a model. Models compose by embedding children:

```go
type Model struct {
    commandBar CommandBar      // Child model
    list       messages.Model  // Child model
    sidebar    Sidebar         // Child model
}
```

Parents route messages to children:

```go
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        if m.focusedComponent == focusCommandBar {
            m.commandBar, cmd = m.commandBar.Update(msg)
        } else {
            m.list, cmd = m.list.Update(msg)
        }
    }
    return m, cmd
}
```

Children handle their own rendering:

```go
func (m Model) View() string {
    return lipgloss.JoinVertical(
        m.list.View(),
        m.commandBar.View(),
    )
}
```

**No coordination at the top.** Each model owns its lifecycle, state, and rendering. Push logic down to where it belongs.

## Update Pattern

Every model's Update follows this structure:

```go
// Rule: return early ONLY if this model is the sole consumer of the message.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    var cmds []tea.Cmd
    var cmd tea.Cmd

    // Handle messages we care about
    switch msg := msg.(type) {
    case childPackage.SomeMsg:
        return m.handleSomeMsg(msg) // sole consumer
    case tea.WindowSizeMsg:
        m = m.updateLayout(msg) // children also need this
    case tea.KeyPressMsg:
        if msg.String() == "tab" {
            m = m.switchFocus()
            return m, nil // sole consumer
        }
    case someSharedMsg:
        m, cmd = m.handleShared(msg) // others also need this
        cmds = append(cmds, cmd)
    }

    // Forward to children - they decide what to handle
    m.child1, cmd = m.child1.Update(msg)
    cmds = append(cmds, cmd)

    m.child2, cmd = m.child2.Update(msg)
    cmds = append(cmds, cmd)

    return m, tea.Batch(cmds...)
}
```

**One rule: return early ONLY if you're the sole consumer.**

- `return` = "I handled this, no one else needs it"
- No return = "I may have acted on this, but keep forwarding"

This prevents the classic bug: "why didn't this component receive this message?" Check the parent's Update—is there an early return intercepting it?

## Focus Management

Only one component receives keyboard input at a time. Chat manages focus:

```go
type focusTarget int

const (
    focusCommandBar focusTarget = iota
    focusMessageList
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if msg.String() == "tab" {
        m = m.switchFocus()
        return m, nil
    }
    
    // Route to focused component
    if m.focusedComponent == focusMessageList {
        m.list, cmd = m.list.Update(msg)
    } else {
        m.commandBar, cmd = m.commandBar.Update(msg)
    }
    return m, cmd
}
```

Tab switches focus. Components only handle messages when focused.

## Chat: The Nerve Center

Chat model (`internal/tui/app/chat/model.go`) handles:
- Message submission and streaming
- Layout switching (empty → with messages)
- Focus between input and message list
- Tool execution coordination

Key state:

```go
type Model struct {
    // Conversation
    conversationID domain.ConversationID
    rawMessages    []domain.Message      // From database
    
    // Streaming
    streaming        *domain.Message     // Current assistant response
    streamingItemIDs []string            // IDs we're updating
    eventCh          chan TurnEvent      // Events from Turn
    
    // Components
    commandBar CommandBar
    list       messages.Model
    sidebar    Sidebar
    
    // Focus
    focusedComponent focusTarget
}
```

## The Turn Abstraction

Turn encapsulates the conversation loop: send → stream → execute tools → repeat.

```go
type TurnEvent struct {
    Message    *domain.Message  // Current streaming state
    ToolResult *domain.Block    // Tool just executed
    Messages   []domain.Message // All messages when done
    Done       bool
    Error      error
}

type Turn interface {
    Run(ctx, conversationID, messages, tools, eventCh chan<- TurnEvent)
}
```

The loop:

```
┌─────────────────────────────────────────────────────────┐
│                      Turn.Run()                         │
│                                                         │
│  ┌──────────────┐                                       │
│  │ Build request│ (messages + tools)                    │
│  └──────┬───────┘                                       │
│         │                                               │
│         ▼                                               │
│  ┌──────────────┐     ┌──────────────┐                  │
│  │ Stream from  │────▶│ Send events  │ (Message)        │
│  │ Chat API     │     │ to channel   │                  │
│  └──────┬───────┘     └──────────────┘                  │
│         │                                               │
│         ▼                                               │
│  ┌──────────────┐                                       │
│  │ Check stop   │                                       │
│  │ reason       │                                       │
│  └──────┬───────┘                                       │
│         │                                               │
│    ┌────┴────┐                                          │
│    │         │                                          │
│    ▼         ▼                                          │
│ "end_turn"  "tool_use"                                  │
│    │         │                                          │
│    │         ▼                                          │
│    │  ┌──────────────┐                                  │
│    │  │ Execute tools│ locally                          │
│    │  │ Send events  │ (ToolResult)                     │
│    │  └──────┬───────┘                                  │
│    │         │                                          │
│    │         └──────────┐                               │
│    │                    │                               │
│    │                    ▼                               │
│    │            ┌──────────────┐                        │
│    │            │ Add results  │                        │
│    │            │ to messages  │                        │
│    │            └──────┬───────┘                        │
│    │                   │                                │
│    │                   └───────────────────▶ loop       │
│    │                                                    │
│    ▼                                                    │
│  ┌──────────────┐                                       │
│  │ Send Done    │ (all messages)                        │
│  │ Close channel│                                       │
│  └──────────────┘                                       │
└─────────────────────────────────────────────────────────┘
```

Turn runs in a goroutine. Chat model waits on the channel:

```go
func (m Model) waitForNextEvent() tea.Cmd {
    return func() tea.Msg {
        event := <-m.eventCh
        return turnEventMsg{event: event}
    }
}
```

## Message Rendering

Messages display in a scrollable list. The list holds Items:

```go
type Item interface {
    ID() string
    Update(msg tea.Msg) (Item, tea.Cmd)
    Render(width int) string
    Height(width int) int
    SetFocused(bool) Item
}
```

Three item types:

**user.Model** — User messages. Simple text rendering.

**assistant.Model** — Assistant messages. Text blocks, may have thinking blocks.

**tools.Model** — Tool calls. Header with status icon, expandable body.

Tool models delegate body rendering to specific implementations:

```go
func New(theme, use, result) *Model {
    m := &Model{Theme: theme, Use: use, Result: result}
    
    switch use.Name {
    case "query":
        m.body = query.New(theme, use, result)
    case "start_journey":
        m.body = startjourney.New(theme, use, result)
    default:
        m.body = generic.New(theme, use, result)
    }
    
    return m
}
```

Each body type renders its tool-specific content (query shows SQL + results table, journey shows status, etc.).

## Tools: Execution vs Rendering

Tools have two parts in different places:

**Execution** (`internal/tui/app/tools/`) — Tool definitions and logic:
```go
type Tool struct {
    DB sqlite.DB
}

func (t *Tool) Definition() chat.Tool {
    return chat.Tool{
        Name:        "query",
        Description: "Execute read-only SQL...",
        InputSchema: schema,
    }
}

func (t *Tool) Execute(input json.RawMessage) (any, error) {
    // Run the query
}
```

**Rendering** (`internal/tui/app/chat/messages/assistant/tools/`) — How tools display:
```go
type Model struct {
    Theme  *styles.Theme
    Use    *domain.ToolUse
    Result *domain.ToolResult
    body   Body
}

func (m *Model) Render(width int) string {
    // Render header + body
}
```

Why separate? Execution is global—MCP uses the same tool definitions. Rendering is TUI-specific.

## Layouts

Two layout wrappers:

**header.Model** — Just a title bar. Used for empty states.

**base.Model** — Title bar + footer with key hints. Used when there's content.

Chat switches between them:

```go
func (m Model) View() string {
    if !m.hasMessages() {
        return m.headerLayout.Render(emptyView)
    }
    return m.baseLayout.Render(messagesView)
}
```

## Async Operations

Bubbletea is single-threaded. Async work happens via Cmd:

```go
// Start async work
func (m Model) handleSubmit() (Model, tea.Cmd) {
    go m.turn.Run(ctx, convID, messages, tools, m.eventCh)
    return m, m.waitForNextEvent()  // Cmd that reads from channel
}

// Handle result
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case turnEventMsg:
        // Process the event
        return m, m.waitForNextEvent()  // Keep listening
    }
}
```

The pattern: spawn goroutine, return Cmd that waits for result, handle result in Update.

## Code Location

```
internal/tui/
├── model.go                 Root (onboarding vs app)
├── app/
│   ├── model.go            App shell
│   ├── chat/
│   │   ├── model.go        Chat page
│   │   ├── turn.go         Conversation loop
│   │   ├── command_bar.go  Input
│   │   ├── sidebar.go      Conversation list
│   │   └── messages/
│   │       ├── list.go     Scrollable viewport
│   │       ├── item.go     Item interface
│   │       ├── user/       User message model
│   │       └── assistant/  Assistant + tools
│   └── tools/              Tool execution (query, journey)
├── onboarding/             Auth flow, account selection
├── components/             Shared components (loader, input, etc.)
└── layouts/                Layout wrappers (header, base)
```
