# TEA Patterns

How we build TUI components with Bubbletea and Lipgloss. Read this before writing any UI code.

## The Layout System

This is the most important section. Most rendering bugs come from layout mistakes.

### The Layout Struct Pattern

Every parent calculates layout once, then passes regions to children. This is how Crush (the reference implementation) does it.

```go
type layout struct {
    area    image.Rectangle  // total available space
    header  image.Rectangle  // region for header
    main    image.Rectangle  // region for main content
    sidebar image.Rectangle  // region for sidebar
    editor  image.Rectangle  // region for editor
    status  image.Rectangle  // region for status bar
}
```

Calculate layout on resize, store it, use it in View:

```go
func (m *Model) SetSize(width, height int) {
    m.width = width
    m.height = height
    m.layout = m.generateLayout(width, height)
    
    // Propagate to children
    m.header.SetWidth(m.layout.header.Dx())
    m.main.SetSize(m.layout.main.Dx(), m.layout.main.Dy())
    m.status.SetWidth(m.layout.status.Dx())
}

func (m *Model) generateLayout(w, h int) layout {
    area := image.Rect(0, 0, w, h)
    
    // Split: status bar at bottom (1 row)
    mainArea, statusArea := splitVertical(area, h-1)
    
    // Split main: sidebar on right (30 cols)
    contentArea, sidebarArea := splitHorizontal(mainArea, w-30)
    
    // Add margins by shrinking rectangles
    contentArea.Min.X += 1  // left padding
    contentArea.Max.X -= 1  // right padding
    contentArea.Max.Y -= 1  // bottom margin before status
    
    return layout{
        area:    area,
        main:    contentArea,
        sidebar: sidebarArea,
        status:  statusArea,
    }
}
```

### Who Owns What

| Concern | Owner |
|---------|-------|
| External margins between siblings | Parent |
| Internal padding within borders | Child (only if it owns a visual box like a border) |
| Width and height | Parent tells child via SetSize |
| Content truncation | Child (using MaxHeight) |

**Children render at column 0.** A child component fills its full width — it never adds leading indentation or margins to position itself within the parent. The parent handles all positioning (via `Padding`, `Place`, or layout calculations). This makes every component embeddable in any context without surprises.

```go
// CORRECT: Child renders flush — parent positions it
func (m *Child) View() string {
    return lipgloss.NewStyle().
        Width(m.width).
        Background(m.theme.Bg).
        Render(m.content)
}

// WRONG: Child adds indentation — assumes parent context
func (m *Child) View() string {
    return lipgloss.NewStyle().
        PaddingLeft(2).   // NO - this is the parent's job
        Render(m.content)
}

// CORRECT: Parent controls spacing between children
func (m *Parent) View() string {
    return lipgloss.JoinVertical(lipgloss.Top,
        m.header.View(),
        "",                // Parent adds the 1-line gap
        m.content.View(),
        m.footer.View(),
    )
}

// WRONG: Child adds external margin
func (m *Child) View() string {
    return lipgloss.NewStyle().
        MarginBottom(1).  // NO - this affects siblings
        Render(m.content)
}
```

### Size Propagation: Top-Down Only

Size flows down from WindowSizeMsg through the component tree:

```
tea.WindowSizeMsg
    │
    ▼
app.Model.SetSize(w, h)
    │
    ├── layout = generateLayout(w, h)
    │
    ├── header.SetWidth(layout.header.Dx())
    │
    ├── content.SetSize(layout.main.Dx(), layout.main.Dy())
    │       │
    │       ├── messagelist.SetSize(...)
    │       └── editor.SetSize(...)
    │
    └── status.SetWidth(layout.status.Dx())
```

Never calculate size in View(). By the time View() runs, every component already knows its dimensions.

### The Two Types of Components

Every component falls into exactly one of two categories:

| Type | Size Interface | Who Decides Height | View Contract |
|------|---------------|-------------------|---------------|
| **Fixed** | `SetWidth(w)` + `Height() int` | Component (from content or constant) | Renders naturally, parent measures |
| **Flexible** | `SetSize(w, h)` | Parent tells child | Renders exactly h lines |

**Fixed-height components** determine their own height. This includes:
- Components with constant height (statusbar = 1, commandbar = 5)
- Content-driven components where height depends on content (text blocks, user messages)

The parent calls `SetWidth(w)`, then asks `Height()` to know how much space the component needs.

```go
// Fixed-height component (constant)
type StatusBar struct {
    width int
}

func (m *StatusBar) SetWidth(width int) {
    m.width = width
}

func (m *StatusBar) Height() int {
    return 1  // Always 1 row
}

func (m *StatusBar) View() string {
    return lipgloss.NewStyle().Width(m.width).Render(m.content)
}
```

```go
// Fixed-height component (content-driven)
type TextBlock struct {
    width    int
    rendered string
}

func (m *TextBlock) SetWidth(width int) {
    m.width = width
    m.rendered = renderMarkdown(m.text, width)
}

func (m *TextBlock) Height() int {
    return lipgloss.Height(m.rendered)
}

func (m *TextBlock) View() string {
    return m.rendered
}
```

**Flexible components** are told their exact dimensions by the parent. They render exactly that size, handling scrolling or truncation internally.

```go
// Flexible component (viewport/scrollable)
type MessageList struct {
    width  int
    height int
    // ... scroll state
}

func (m *MessageList) SetSize(width, height int) {
    m.width = width
    m.height = height
}

func (m *MessageList) View() string {
    content := m.renderVisibleContent()
    return lipgloss.NewStyle().
        Width(m.width).
        Height(m.height).
        MaxHeight(m.height).  // Ensure no overflow
        Render(content)
}
```

### How Parents Use These Types

```go
func (m *Parent) SetSize(width, height int) {
    m.width = width
    m.height = height
    
    // Fixed components: set width, ask for height
    m.statusbar.SetWidth(width)
    m.commandbar.SetWidth(width)
    
    // Calculate remaining space for flexible component
    fixedHeight := m.statusbar.Height() + m.commandbar.Height()
    flexibleHeight := height - fixedHeight
    
    // Flexible component: tell it exactly what size to be
    m.messagelist.SetSize(width, flexibleHeight)
}

func (m *Parent) View() string {
    return lipgloss.JoinVertical(lipgloss.Top,
        m.statusbar.View(),
        m.messagelist.View(),
        m.commandbar.View(),
    )
}
```

### Height and MaxHeight

Lipgloss `Height()` sets a **minimum**. It pads short content but does NOT truncate long content.

```go
// Height(10) with 20 lines of content = 20 lines rendered (overflow!)
style := lipgloss.NewStyle().Height(10)
rendered := style.Render(twentyLines) // Still 20 lines

// MaxHeight(10) truncates to 10 lines
style := lipgloss.NewStyle().Height(10).MaxHeight(10)
rendered := style.Render(twentyLines) // Now 10 lines
```

Always use both when you need exact dimensions:

```go
func (m *Model) View() string {
    return lipgloss.NewStyle().
        Width(m.width).
        Height(m.height).
        MaxWidth(m.width).
        MaxHeight(m.height).
        Render(m.content)
}
```

### Composition with JoinVertical/JoinHorizontal

Use lipgloss join functions, not manual `\n` concatenation:

```go
// CORRECT
view := lipgloss.JoinVertical(lipgloss.Top,
    header,
    content,
    footer,
)

// WRONG - alignment issues, hard to debug
view := header + "\n" + content + "\n" + footer
```

Join functions handle:
- Alignment (Top/Bottom/Center for vertical, Left/Right/Center for horizontal)
- Padding shorter elements to match the longest
- Proper newline handling

### Gaps and Spacing

Empty strings in JoinVertical create 1-line gaps:

```go
lipgloss.JoinVertical(lipgloss.Top,
    header,
    "",       // 1-line gap
    content,
    "",       // 1-line gap  
    footer,
)
```

Account for gaps in layout calculations:

```go
func (m *Model) generateLayout(w, h int) layout {
    const (
        headerHeight     = 1
        gapAfterHeader   = 1
        footerHeight     = 1
        gapBeforeFooter  = 1
    )
    
    contentHeight := h - headerHeight - gapAfterHeader - gapBeforeFooter - footerHeight
    // ...
}
```

### Layers and Overlays (Dialogs, Toasts)

For overlays that float above content, use lipgloss.Layer and Compositor:

```go
func (m *Model) View() string {
    // Base layer
    base := lipgloss.JoinVertical(lipgloss.Top, 
        m.content.View(),
        m.status.View(),
    )
    
    layers := []*lipgloss.Layer{
        lipgloss.NewLayer(base),
    }
    
    // Dialog overlay (positioned absolutely)
    if m.dialog != nil {
        dialogView := m.dialog.View()
        w, h := lipgloss.Width(dialogView), lipgloss.Height(dialogView)
        x := (m.width - w) / 2   // Center horizontally
        y := (m.height - h) / 2  // Center vertically
        
        layers = append(layers,
            lipgloss.NewLayer(dialogView).X(x).Y(y).Z(1),
        )
    }
    
    return lipgloss.NewCompositor(layers...).Render()
}
```

## Models

Everything is a model. Models compose models.

```
app.go
  ├── statusbar/
  ├── toast/
  ├── onboarding/
  │   └── sync/
  └── chat/
      ├── editor/
      └── messagelist/
```

A model:
- Has state (width, height, data, child models)
- Implements `Init()`, `Update()`, `View()`
- Receives dependencies via constructor
- Receives size via `SetSize()` or `SetWidth()`

### Construction vs Sizing

**Never pass dimensions to constructors.** Construction and sizing are separate concerns:

- `New()` receives only dependencies (theme, db, services, etc.)
- `SetSize()` or `SetWidth()` is called by the parent after construction

This ensures components work correctly regardless of when they're created relative to `WindowSizeMsg`.

```go
// WRONG - dimensions in constructor
func New(theme *styles.Theme, width, height int) *Model {
    ta := textarea.New()
    ta.SetWidth(width)  // width could be 0 if WindowSizeMsg hasn't arrived!
    return &Model{width: width, height: height}
}

// RIGHT - only dependencies in constructor
func New(theme *styles.Theme) *Model {
    return &Model{
        theme: theme,
        ta:    textarea.New(),
    }
}

// Parent sizes after construction
child := New(theme)
child.SetWidth(width)  // Called when parent knows its size
```

### Model Structure

```go
type Model struct {
    // Dependencies (injected via New)
    theme  *styles.Theme
    db     sqlite.DB
    
    // Dimensions (set via SetSize, start at zero)
    width  int
    height int
    layout layout
    
    // State
    items []Item
    
    // Children
    list *list.Model
}

func New(theme *styles.Theme, db sqlite.DB) *Model {
    return &Model{
        theme: theme,
        db:    db,
        list:  list.New(theme),
    }
}
```

## Theming and Backgrounds

Terminals have no transparency. Every character cell has an explicit background color. There are no layers, no inheritance, no CSS-like cascading. If a cell doesn't have a background escape sequence, the terminal default punches through.

Two rules:

1. **Every styled text sets `Background(theme.Bg)`.** Always. No exceptions.
2. **The parent decides the background.** Children just use `theme.Bg`.

### Rule 1: Always Set Background

```go
// WRONG - terminal default bg punches through
lipgloss.NewStyle().Foreground(theme.Text).Render("hello")

// RIGHT - explicit bg on every style
lipgloss.NewStyle().Foreground(theme.Text).Background(theme.Bg).Render("hello")
```

This applies everywhere: components, inline styles, pre-built theme styles, markdown renderers, gradients. If it renders visible text, it sets `Background(theme.Bg)`.

The theme's pre-built styles (`theme.Styles.Body`, `theme.Styles.Title`, etc.) already include the background. Use them directly:

```go
theme.Styles.Body.Render("some text")    // bg is already set
theme.Styles.Title.Render("heading")     // bg is already set
```

### Rule 2: Parent Decides Background

**Children never decide their own background.** They just use `theme.Bg`. The parent decides what color that is by passing the right theme.

To change a child's background: `theme.WithBg(color)`. One line. That's it.

`WithBg()` returns a new theme copy with `Bg` set to the given color and all `Styles` rebuilt. The child doesn't know or care what surface it's on.

```go
// Parent decides which children get which background
func (m *Model) ensureBlocks() {
    // Text blocks: normal page background
    m.textBlock = text.New(m.theme, m.width)

    // Tools: elevated surface
    elevated := m.theme.WithBg(m.theme.BgElevated)
    m.toolBlock = tool.New(elevated, m.width)
}

// Child just uses theme.Bg — doesn't know what color it is
func (m *ToolBlock) View() string {
    return lipgloss.NewStyle().
        Background(m.theme.Bg).   // page bg? elevated? doesn't matter
        Foreground(m.theme.Text).
        Width(m.width).
        Render(m.content)
}
```

This is the **only** way to change a child's background. Never:
- Remove backgrounds to make something "transparent"
- Hard-code colors in children
- Hack renderers (glamour, gradients) to change backgrounds
- Modify theme tokens to fix one component's color

Just pass the right theme.

## Messages

State flows down (constructors, setters). Events flow up (messages).

### Emitting Messages

When something happens that other components need to know:

```go
// Define in internal/app/msgs/
type SyncStateChanged struct {
    State powersync.State
}

// Emit from the component that detects the change
func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg.(type) {
    case pollMsg:
        if m.stateChanged() {
            return func() tea.Msg {
                return msgs.SyncStateChanged{State: m.currentState}
            }
        }
    }
    return nil
}
```

### Listening to Messages

Any model can listen to any message - they broadcast to the whole tree:

```go
func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case msgs.SyncStateChanged:
        if msg.State == powersync.Ready {
            // React to sync completing
        }
    }
    return nil
}
```

### Message Organization

```
internal/app/msgs/
  ├── toast.go    # ShowError, ShowWarning, ShowSuccess
  ├── sync.go     # SyncStateChanged
  └── nav.go      # NavigateTo, GoBack
```

## Forwarding Messages

Parents forward messages to children:

```go
func (m *Model) Update(msg tea.Msg) tea.Cmd {
    var cmds []tea.Cmd
    
    // Components that listen to everything
    cmds = append(cmds, m.toast.Update(msg))
    cmds = append(cmds, m.statusbar.Update(msg))
    
    // Forward to active page only
    switch m.state {
    case stateChat:
        cmds = append(cmds, m.chat.Update(msg))
    case stateOnboarding:
        cmds = append(cmds, m.onboarding.Update(msg))
    }
    
    return tea.Batch(cmds...)
}
```

## Polling

For background state checks:

```go
const pollInterval = 500 * time.Millisecond

type pollMsg struct{}

func (m *Model) Init() tea.Cmd {
    return m.poll()
}

func (m *Model) poll() tea.Cmd {
    return tea.Tick(pollInterval, func(time.Time) tea.Msg {
        return pollMsg{}
    })
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg.(type) {
    case pollMsg:
        // Check state, emit messages if changed
        return m.poll()  // Continue polling
    }
    return nil
}
```

Only one component should poll a given resource. Others listen to messages.

## Anti-Patterns

### Don't pass dimensions to constructors

```go
// WRONG - component may be created before WindowSizeMsg
func New(theme *styles.Theme, width int) *Model {
    ta.SetWidth(width)  // width could be 0!
    return &Model{width: width}
}

// RIGHT - only dependencies in constructor
func New(theme *styles.Theme) *Model {
    return &Model{theme: theme}
}
```

### Don't calculate size in View

```go
// WRONG
func (m *Model) View() string {
    width := m.parentWidth - 4  // Calculating in View
    return lipgloss.NewStyle().Width(width).Render(...)
}

// RIGHT
func (m *Model) View() string {
    return lipgloss.NewStyle().Width(m.width).Render(...)  // Already calculated
}
```

### Don't use magic numbers

```go
// WRONG
contentHeight := height - 5  // What is 5?

// RIGHT  
const (
    headerHeight = 1
    statusHeight = 1
    editorHeight = 3
)
contentHeight := height - headerHeight - statusHeight - editorHeight
```

### Don't add external margins in components

```go
// WRONG - affects layout of siblings
func (m *Model) View() string {
    return lipgloss.NewStyle().MarginTop(1).Render(...)
}

// RIGHT - parent controls gaps
// (parent's View)
return lipgloss.JoinVertical(lipgloss.Top, prev, "", m.View())
```

### Don't forget Background on styled text

```go
// WRONG - terminal default bg punches through
lipgloss.NewStyle().Foreground(theme.Accent).Render("label")

// RIGHT - always set bg when you have a theme
lipgloss.NewStyle().Foreground(theme.Accent).Background(theme.Bg).Render("label")
```

### Don't use Height without MaxHeight for flexible content

```go
// WRONG - content can overflow
style := lipgloss.NewStyle().Height(10)

// RIGHT - content is truncated
style := lipgloss.NewStyle().Height(10).MaxHeight(10)
```

## Adding a New Component

1. Create `internal/app/mycomponent/mycomponent.go`
2. Define Model struct with dependencies, dimensions, state, children
3. Implement `New(deps)` constructor
4. Implement `Init() tea.Cmd`
5. Implement `Update(msg tea.Msg) tea.Cmd`
6. Implement `View() string` — set `Background(theme.Bg)` on every style
7. Implement `SetSize(w, h)` or `SetWidth(w)` + `Height() int`
8. Add message types to `internal/app/msgs/` if it emits events
9. Wire into parent's constructor, Update (forward messages), View (compose), and SetSize (propagate)

## Quick Reference

| Function | Use For |
|----------|---------|
| `lipgloss.JoinVertical(pos, ...)` | Stack components vertically |
| `lipgloss.JoinHorizontal(pos, ...)` | Place components side by side |
| `lipgloss.Place(w, h, hPos, vPos, str)` | Center content in whitespace |
| `lipgloss.Width(str)` | Measure rendered width |
| `lipgloss.Height(str)` | Measure rendered height |
| `lipgloss.NewLayer(str).X(x).Y(y)` | Create positioned layer |
| `lipgloss.NewCompositor(layers...)` | Compose layers for overlays |
| `style.Width(n).Height(n)` | Set minimum dimensions |
| `style.MaxWidth(n).MaxHeight(n)` | Set maximum dimensions (truncate) |
| `style.Padding(t, r, b, l)` | Internal spacing |
| `style.Margin(t, r, b, l)` | External spacing (use sparingly) |
