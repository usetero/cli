# Architecture

How the CLI is built and how the pieces fit together.

This document assumes you understand what Tero is and how the product works. If you haven't read the [product doc](../../knowledge-base/product.md), start there. This doc focuses on implementation—how we build the CLI, not what it does or why.

---

## What This Is

The CLI is a presentation layer for the Tero control plane. It provides three interfaces: a TUI for interactive use, an MCP server for coding agents, and traditional commands for scripting. All three share the same foundation—authentication, data access, and the app layer—but present information differently.

The CLI never implements intelligence. It doesn't analyze log patterns, calculate waste percentages, or make decisions about data quality. When you're tempted to add logic that feels smart—stop. That belongs in the control plane.

---

## How Data Flows

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              CLI                                        │
│  ┌───────────┐    ┌───────────┐    ┌───────────┐                       │
│  │    TUI    │    │    MCP    │    │ Commands  │                       │
│  └─────┬─────┘    └─────┬─────┘    └─────┬─────┘                       │
│        └────────────────┼────────────────┘                              │
│                         ▼                                               │
│                 ┌───────────────┐                                       │
│                 │   App Layer   │  Services, interfaces, domain logic  │
│                 └───────┬───────┘                                       │
│                         ▼                                               │
│                 ┌───────────────┐                                       │
│                 │ Local SQLite  │  Catalog, messages, context, views   │
│                 └───────┬───────┘                                       │
└─────────────────────────┼───────────────────────────────────────────────┘
                          │
                          │ PowerSync (bidirectional sync)
                          ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Control Plane                                   │
│  ┌───────────┐    ┌───────────┐    ┌───────────┐                       │
│  │  GraphQL  │    │   Chat    │    │  Catalog  │                       │
│  │    API    │    │  Service  │    │  Service  │                       │
│  └───────────┘    └─────┬─────┘    └───────────┘                       │
│                         │                                               │
│                         ▼                                               │
│                 ┌───────────────┐                                       │
│                 │    Claude     │  AI reasoning, tool calls            │
│                 └───────────────┘                                       │
└─────────────────────────────────────────────────────────────────────────┘
```

**The chat loop:**

1. User types a message in the command bar
2. Client writes user message to local SQLite
3. PowerSync syncs the message to the control plane
4. Control plane loads conversation history, builds system prompt, calls Claude
5. Claude responds (possibly with tool calls for views, context changes, etc.)
6. Control plane writes assistant message to database
7. PowerSync syncs the response back to the client
8. Client renders the new message

The client never calls Claude directly. It writes messages and renders responses. The control plane handles everything in between.

---

## The App Layer

Between the interfaces (TUI, MCP server) and the data sources sits the app layer. It orchestrates authentication, manages preferences, and provides domain operations.

### Three Tiers

**Services** own domain concepts. `AuthService` knows about device authorization flows and token refresh. `PreferencesService` knows about email addresses and organization IDs. Services don't know about YAML files, OS keychains, or GraphQL queries—they work with interfaces.

**Interfaces** are defined by consumers, not providers. `PreferencesService` needs key-value storage, so it depends on `Store`—a simple interface with `Get`, `Set`, `Save`. It doesn't care whether the implementation uses YAML or a database. Interfaces live in the app package alongside services.

**Implementations** handle the messy details. `config.Config` implements `Store` with YAML files. `keyring.Keyring` implements `SecureStorage` with OS keychains. The generated GraphQL client implements `APIClient`. Implementations can be swapped without touching services.

### Wiring

Composition happens in `cmd/`. That's where implementations get wired to interfaces—config store created, keyring initialized, clients configured, TUI launched with all its dependencies. Nowhere else does this wiring happen.

### Testing

The app layer includes `apptest`—mock implementations for every interface. These mocks use function fields, making them trivial to configure. Need to test what happens when authentication fails? Set `MockOAuthProvider.PollAuthenticationFunc` to return an error. See [TESTING.md](TESTING.md) for details.

---

## The TUI

The TUI is built on [Bubbletea](https://github.com/charmbracelet/bubbletea), a Go framework based on The Elm Architecture. Your application is a model (state), an update function (how state changes), and a view function (how to render state). Messages flow in, trigger updates, the model changes, the view re-renders.

### Two Modes

The TUI has two modes: **onboarding** and **app**. They're fundamentally different experiences.

**Onboarding** is a linear flow. Steps chain together—auth, role selection, organization setup, account creation. Each step knows what comes next. A Flow orchestrator manages transitions. When the final step completes, the TUI transitions to app mode. Onboarding uses FlowContext to accumulate data as steps complete.

**App mode** is the main experience. Chat is home—you always start there and can always escape back. You can focus on other pages (services, policies, expanded views) but you never lose access to the command bar or the ability to return.

### Value vs Pointer Receivers

Bubbletea follows Elm's pattern where `Update()` returns the new state via value receivers. This works well for top-level models (pages, modes) where you want explicit state transitions.

But for leaf components (footer, header, text input), value receivers create a footgun: forgetting to capture the return value silently loses state. So our components use pointer receivers—standard Go for methods that mutate.

The mental model: **pages and modes use value receivers** (Elm-style), **components use pointer receivers** (Go-style).

---

## The App Shell

The app shell (`internal/tui/app/`) orchestrates everything in app mode.

### What App Owns

The app shell owns the persistent UI elements that stay visible regardless of what page you're focused on:

- **Command bar** — always at the bottom, always accepts input
- **Navigation state** — what's focused, breadcrumb trail back to chat
- **Chrome** — sidebar (wide) or header (narrow), footer with keybindings

Pages don't know about the command bar. They render their content, and the app shell composes everything together.

### Layout

Two layouts depending on terminal width:

```
Wide (120+ chars):                    Narrow:
┌──────────────────┬─────────┐        ┌──────────────────────────┐
│                  │         │        │ Header                   │
│                  │ Sidebar │        ├──────────────────────────┤
│  Focused Page    │         │        │                          │
│                  │         │        │     Focused Page         │
│                  │         │        │                          │
├──────────────────┴─────────┤        ├──────────────────────────┤
│ > command bar              │        │ > command bar            │
├────────────────────────────┤        ├──────────────────────────┤
│ footer (keybindings)       │        │ footer                   │
└────────────────────────────┘        └──────────────────────────┘
```

The command bar is always visible. The sidebar shows context (active entities, metadata). On narrow terminals, the header shows condensed info.

### Focus and Navigation

Chat is always home. When you navigate to another page (via `/services`, expanding a view, etc.), that page becomes focused. Escape returns focus to the previous page, eventually back to chat.

```
Navigation: Chat → Services → Service Detail
Escape:     Chat ← Services ← Service Detail
```

How the focused page is *presented* is flexible—it could replace the content area entirely, split the view, or overlay. The conceptual model is focus, not a rigid UI pattern. What matters:

1. Command bar stays visible
2. You can always escape back to chat
3. Breadcrumbs show where you are ("Services · Waste View ← from chat")

### Chrome

The app shell updates chrome based on the focused page's declarations:

- `Title()` — displayed in header/breadcrumbs
- `Metadata()` — displayed in sidebar, sorted by priority
- `KeyBindings()` — displayed in footer
- `Commands()` — slash commands this page supports
- `AcceptsNaturalLanguage()` — whether input goes to AI

---

## Chat

Chat is the home screen. It displays the conversation history.

### The Command Bar

The command bar lives in the app shell, not in chat. It's multi-modal:

- **Natural language** → goes to the AI
- **Slash commands** (`/services`) → navigate to pages
- **@ references** (`@checkout-service`) → add entities to context

The interface detects what you're typing and adapts. Slash commands show completions. @ references show entity search. Plain text goes to the AI.

Because the command bar is at the app level, you can type commands even when focused on the services page. "Sort by cost" while viewing the services table goes to the AI and adjusts the view.

### Messages

Messages load from local SQLite and re-render when data changes. Chat subscribes to table changes via PowerSync. When the `messages` table changes, it refreshes.

Message content is an array of blocks:

- **text** — prose content, rendered as markdown
- **thinking** — AI reasoning, rendered as collapsible
- **tool_use** — AI calling a tool, rendered as status indicator
- **tool_result** — tool output, rendered based on tool type

Views are tool results. When the AI creates a view, it's a tool call with a view spec. The client executes the query and renders inline.

### Sending Messages

When the user submits:

1. Client writes user message to local SQLite
2. PowerSync syncs to control plane
3. Control plane processes (calls Claude, etc.)
4. Assistant message syncs back
5. Client renders it

The client doesn't wait for a response. It writes optimistically and renders when data arrives via sync.

---

## Views

Views are how data gets displayed. A view is a configured query: columns, sort, filters, grouping.

### Three Data Shapes

All views render one of three shapes:

- **Tables** — rows and columns (most catalog pages)
- **Categorical** — numbers with labels (pie charts, bar charts)
- **Time series** — values over time (line charts, trends)

### Inline vs Focused

Views appear in two contexts:

- **Inline** in chat — compact, part of the message flow, limited interactivity
- **Focused** as the active page — full controls, complete interactivity

When you expand an inline view, that view's page becomes focused. The command bar stays. You can chat with the AI to modify the view. Escape returns to chat.

### View State

Views exist in three states:

- **Default** — what you see when you type `/services`
- **Saved** — pinned by you or teammates, persisted
- **Ephemeral** — AI-generated during conversations, tied to that session

Ephemeral views have version history. As you modify a view during a detour, versions accumulate. The inline view in chat reflects the latest version.

### How the AI Creates Views

The AI sends view configurations, not data. The client runs the query against local SQLite.

```
AI tool call:
{
  "tool": "show_table",
  "input": {
    "entity": "services",
    "columns": ["name", "waste", "cost"],
    "sort": {"field": "waste", "desc": true},
    "filter": {"team": "payments"},
    "limit": 10
  }
}

Client:
1. Receives tool call via message sync
2. Executes query against local SQLite
3. Renders table inline in chat
```

This is fast because the catalog is local. The AI doesn't need to send data or wait for queries.

---

## Data

### Local SQLite

The Master Catalog syncs to local SQLite via PowerSync. Every service, log event, metric, trace span, policy—always up to date, always queryable locally.

This is why navigation is instant. Filtering, sorting, searching—all local queries. No network round trips.

### What's Local vs Remote

**Local (via PowerSync):**
- Catalog entities (services, log events, policies, etc.)
- Conversations and messages
- Conversation context
- Views

**Remote (via API):**
- Actual telemetry (raw logs, metric data points)
- Actions requiring immediate server processing
- Authentication flows

The catalog describes data. The telemetry is the actual data. When you need real logs, Tero proxies the request to the vendor.

### Context

Context is stored in `conversation_contexts`—a join table tying catalog entities to conversations. When you @ reference something:

1. Client writes to `conversation_contexts` locally
2. PowerSync syncs to control plane
3. Control plane sees the context on next turn
4. System prompt includes full entity data

The client shows what's in context (sidebar on wide terminals, overlay on narrow). The AI sees everything in context on every turn.

---

## Patterns

### Pages Decide, Components Render

Components don't make layout decisions. They accept parameters and render. A header doesn't decide where to position itself. Pages pass that information down.

### Stateless by Default

Prefer stateless components. A component that takes parameters and returns output is easier to understand than one managing internal state. Some components need state (text inputs manage cursor position)—keep it minimal.

### Dimensions Flow Down

Pages receive width and height via `SetSize()`. They do layout math, then pass constrained dimensions to children. No component reaches up to ask "how big am I?"

### Cursor Positioning with Markers

Don't calculate cursor X/Y manually. Embed a marker (`CursorMarker`) in the view string where the cursor should appear. The TUI extracts it from the final rendered output and calculates position.

### Configuration

All CLI config goes in `CLIConfig` (`internal/config/cli.go`). Never read `os.Getenv()` elsewhere.

---

## The MCP Server

`tero mcp` runs the CLI as an MCP server for coding agents.

The MCP server translates tool calls into queries—GraphQL for the control plane, local SQLite for catalog data—and returns structured responses. No UI state, no rendering. Just protocol translation.

---

## Further Reading

- [Product doc](../../knowledge-base/product.md) — what we're building and why
- [TESTING.md](TESTING.md) — how to write tests
- [LOGGING.md](LOGGING.md) — how to write logs
- [Control plane chat doc](../../control-plane/docs/chat.md) — how chat works server-side
