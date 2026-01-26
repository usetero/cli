# Architecture

How the CLI is built, how the pieces fit together, and why.

- [What This Is](#what-this-is)
- [How Data Flows](#how-data-flows)
- [The App Layer](#the-app-layer)
- [The TUI](#the-tui)
- [Patterns and Conventions](#patterns-and-conventions)
- [The MCP Server](#the-mcp-server)

---

## What This Is

Tero is a control plane for observability. The CLI is how you interact with it.

The control plane does the hard work—analyzing telemetry, building the Master Catalog, generating policies, enforcing fixes. The CLI's job is to make that intelligence accessible through a beautiful, fast terminal interface. The CLI never implements intelligence. It doesn't analyze log patterns, calculate waste percentages, or make decisions about data quality. When you're working on the CLI and you're tempted to add logic that feels smart—stop. That logic belongs in the control plane.

The CLI provides three interfaces: a TUI for interactive use, an MCP server for coding agents, and eventually traditional commands for scripting. All three share the same foundation—authentication, data access, and the app layer—but present information differently.

### Chat Is the Canvas

The TUI is a conversational application. Chat is the foundation everything else builds on. You start in chat, navigate deeper into things—a services table, a visualization, a policy review—and you always come back to chat. It's home.

This isn't a dashboard with a chat widget bolted on. Chat is the primary navigation mechanism. You type natural language to explore, slash commands to jump to specific pages, @ references to pull entities into context. All in the same input. The interface adapts based on what you type.

Pages and views still exist—services tables, policy lists, time series charts—and they're fully interactive. But you reach them through chat (or slash commands), and you always escape back. Navigating deeper never feels like you've lost your place. Mental friction kills flow, so we use modals and overlays instead of hard page transitions. You go deeper, do what you need, press escape, and you're back in chat where you left off.

### The Catalog Is Local

The Master Catalog syncs to a local SQLite database on the client via PowerSync. Every service, log event, metric, trace span, policy, and their relationships—always up to date, always queryable locally.

This is the foundation for speed. Filtering services, browsing log events, searching policies—you're querying local data. No network round trips, no loading spinners, no pagination limits. When exploration is free, people explore more.

The actual telemetry—raw logs, metric data points, trace data—stays where it lives. Datadog, Splunk, ClickHouse, wherever. When you need real data, Tero proxies the request through its API. Two data layers: the catalog is local and instant, telemetry is on-demand and streamed.

## How Data Flows

The CLI communicates with the control plane through three channels, each serving a different purpose.

### Chat Protocol

Conversations go through a streaming chat protocol. The client sends messages and receives structured instructions back—run a local query, display a visualization, take an action. The client executes. Intelligence lives entirely in the control plane. The client is a runtime, not a decision-maker.

Chat instructions reference data the client already has locally. When the AI says "show a table of services sorted by waste," the client knows how to query its local SQLite database, build the table, and render it. The AI doesn't send the data—it sends the intent. This means visualizations are interactive from the moment they appear. The user can filter, sort, expand time ranges, adjust columns—because the data is right there.

### Local Database

PowerSync keeps the local SQLite database in sync with the control plane. The sync is continuous and automatic—changes in the control plane propagate to every connected client. The client reads from SQLite for all catalog operations: listing services, browsing log events, querying policies.

PowerSync also supports local writes. Some mutations go through PowerSync—the client writes locally, changes queue and sync to the control plane. Others go through GraphQL directly. The choice depends on what fits the operation. Chat doesn't care about the mechanism. It sends abstract action instructions, and the client decides how to execute.

### GraphQL API

GraphQL handles everything that isn't chat or synced data. Authentication, account management, actions that need immediate server-side processing. Mutations like approving a policy or configuring an integration go through GraphQL. The client sends queries and mutations with an auth token, the control plane processes them and returns structured responses.

### Three Data Shapes

Everything the client displays from local queries falls into three shapes:

**Tables.** Rows and columns. The most common shape for catalog pages—services with their metadata, log events with their properties, policies with their status. Tables are filterable, sortable, and pageable locally.

**Categorical data.** Numbers with labels—for pie charts, bar charts, breakdowns. "45% debug logs, 30% health checks, 25% other." Same underlying data as a table, just rendered differently.

**Time series.** Values over time—for line charts, area charts, trends. Error rates over the last week, waste reduction over a quarter. The time axis makes these distinct from categorical data.

These are the rendering primitives. Every visualization the client displays is one of these three shapes, derived from a query the client runs locally. The AI creates visualizations as shortcuts while you work through a problem, but the user has the same data and controls. The AI just gets you there faster.

## The App Layer

Between the interfaces (TUI, MCP server) and the data sources sits the app layer—services that orchestrate authentication, manage preferences, and provide domain operations the interfaces can work with.

The app layer has three tiers: services define domain operations, interfaces abstract dependencies, and implementations provide the concrete machinery.

### Services

Services own domain concepts and orchestrate operations. `PreferencesService` knows about email addresses and organization IDs. `AuthService` knows about device authorization flows and token refresh. `ServiceService` knows how to list services, check discovery status, and retrieve counts.

Services don't know about YAML files, OS keychains, or GraphQL queries. They work with interfaces—`Store` for preferences, `SecureStorage` for tokens, `OAuthProvider` for authentication, `APIClient` for the control plane. This keeps them focused on domain logic without coupling to implementation details.

Method names follow consistent patterns: `ListByX` for collections, `GetByX` for single items, `CountByX` for totals. Services are scoped by account type where relevant—`ListByDatadogAccount` makes it clear we're querying Datadog-discovered services. When we add Splunk support, we'll add `ListBySplunkAccount` alongside it.

### Interfaces

Interfaces are defined by their consumers, not their providers. `PreferencesService` needs generic key-value storage, so it depends on `Store`—a simple interface with `Get`, `Set`, `GetBool`, `GetList`, and `Save`. It doesn't care whether the implementation uses YAML, JSON, or a database.

This consumer-driven approach keeps interfaces small and focused. `SecureStorage` is just `Get`, `Set`, and `Delete`. `OAuthProvider` defines the device authorization flow without mentioning WorkOS specifically. New implementations can plug in without changing the services that depend on them.

Interfaces live in the app package alongside services. This makes dependencies explicit and keeps the architecture navigable. When you read a service, you see exactly what it needs. When you implement an interface, you know exactly what contract you're fulfilling.

### Implementations

Concrete implementations handle the messy details. `config.Config` implements `Store` with YAML files. `keyring.Keyring` implements `SecureStorage` with OS keychains—Keychain on macOS, Credential Manager on Windows, Secret Service on Linux. `workos.Client` implements `OAuthProvider` with WorkOS API calls. The generated GraphQL client implements `APIClient`.

Implementations can be swapped without touching services. This tier handles all platform-specific concerns, external API integration, file formats, and persistence strategies. Services stay clean. Interfaces stay simple. Implementations deal with reality.

### Testing

The app layer includes `apptest`—a package of mock implementations for every interface. These mocks use function fields, making them trivial to configure in tests. Need to test what happens when authentication fails? Set `MockOAuthProvider.PollAuthenticationFunc` to return an error. Want to verify a service saves preferences correctly? Check what `MockStore.SetFunc` was called with.

This makes tests fast, focused, and deterministic. No file I/O, no network calls, no OS dependencies. Just pure logic tests with complete control over dependencies. See [TESTING.md](TESTING.md) for the full testing philosophy.

## The TUI

The TUI is where most of the CLI's complexity lives. It's built on [Bubbletea](https://github.com/charmbracelet/bubbletea), a Go framework based on The Elm Architecture.

### The Elm Architecture

The pattern is simple: your application is a model (all your state), an update function (how state changes), and a view function (how to render state). Messages flow in—keypresses, API responses, timer ticks—and trigger updates. The model changes, the view re-renders, the cycle continues.

This matters because the TUI has a lot of state to manage. Which page are you on? What's the conversation history? Where's the cursor? What's focused? The Elm Architecture keeps it organized—all state lives in the model, all changes happen in update functions, and views are always pure functions of state.

The pattern is also composable. Pages are models with update and view functions. Components are models with update and view functions. The same pattern repeats at every level. Once you understand it in one place, you understand it everywhere.

One detail worth understanding: Elm enforces immutability—every update returns a new value. Bubbletea follows this with value receivers, where `Update()` returns the new state. This works well for top-level models (pages, modes) where you want state transitions to be explicit.

But Go isn't Elm. For leaf components (footer, header, text input), the value receiver pattern creates a footgun: forgetting to capture the return value silently loses state changes. So our components use pointer receivers instead—standard Go for methods that mutate. This is what Crush (Charm's production app) does. They learned this building a real application.

The mental model: pages and modes are Elm-style (explicit state transitions via value receivers), components are Go-style (direct mutation via pointer receivers).

### Two Modes

The TUI has two modes: onboarding and app. They're fundamentally different experiences with different structures.

**Onboarding** is a linear flow. Steps chain together—auth, role selection, organization setup, account creation, integration configuration. Each step knows what comes next via `Next()`. A Flow orchestrator manages transitions: forward messages to the current step, check if it's complete, advance to the next. When the final step completes, the TUI transitions to app mode.

Onboarding uses FlowContext—a mutable bag that accumulates data as steps complete. The email you enter in auth reaches the organization step. The organization you select reaches account creation. Each step adds its result and passes the context forward. When onboarding finishes, the accumulated data gets sent to the control plane and the context is discarded.

**App mode** is the main experience. Chat is the base layer, always present. The app shell composes chat with chrome—a sidebar (on wide terminals) or a compact header (on narrow ones), plus a command bar at the bottom for input.

Pages in app mode don't chain linearly. They layer. Chat is always the base. Other pages—services table, policy review, visualization detail—appear as modals or overlays on top of chat. You navigate deeper by opening pages, and you escape back to chat. The navigation is a stack: push a page on, interact with it, pop it off. Chat is always at the bottom.

This is different from onboarding's chain pattern, and intentionally so. Onboarding is a one-time, step-by-step process. The app is an open-ended workspace where you move fluidly between chat and focused views.

### The App Shell

The app shell (`internal/tui/app/`) orchestrates everything in app mode. It manages:

**Layout.** Two modes depending on terminal width. Wide terminals (120+ characters) get a sidebar showing context—active entities, session info, metadata. Narrow terminals get a compact header instead. The command bar sits at the bottom in both layouts, always visible.

**Pages.** The chat page is always the base. Other pages stack on top as modals. The app shell routes keyboard input to the right place—the topmost page gets input, escape pops it off.

**Command bar.** The unified input at the bottom. Natural language starts or continues a chat. Slash commands (`/services`, `/policies`) navigate to pages. @ references (`@checkout-service`) pull entities into context. The command bar detects what you're typing and adapts.

**Context sidebar.** As you reference entities in chat—services, log events, policies—they accumulate in the sidebar as active context. The AI sees this context. You can add or remove items, but mostly it builds organically through conversation. On narrow terminals, this context is accessible through a details overlay instead of a persistent sidebar.

### Chat

The chat page is the heart of the application. It displays the conversation—user messages and AI responses—with support for inline visualizations, action prompts, and rich content.

Messages arrive from the chat protocol as a stream. User messages appear immediately (optimistic). AI responses stream in as structured instructions—text to display, queries to run, views to render, actions to offer. The chat page interprets these instructions and builds the conversation display.

Inline visualizations appear directly in the chat flow. A small table, a summary chart—things that make sense at chat scale. Any visualization can be expanded into a full-size modal with all interactive controls. The full-size view has the same data and the same interactivity, just more room to work with.

Sessions are focused—each chat is a session on a specific problem. When you open Tero, you start a new session by default. You can resume previous sessions or branch from them by @-referencing them.

### Pages

Pages beyond chat follow a common interface. Each page declares its title, metadata (for the sidebar), keyboard shortcuts, and whether it accepts natural language input. The app shell uses these declarations to compose the chrome correctly—different pages surface different metadata in the sidebar, different shortcuts in the command bar footer.

Pages receive dimensions from the app shell and render within those bounds. They don't know about the sidebar, the command bar, or other pages. They render their content, report their state, and let the app shell handle composition.

Some pages are catalog views—services, log events, policies. These query local SQLite data and render as interactive tables. Users can filter, sort, adjust columns, and drill into individual entities. These are the same views the AI creates during chat, just accessed directly via slash commands.

Other pages are detail views—a single service, a specific policy, a visualization at full size. These typically appear as modals over chat, though they could also be navigated to directly.

## Patterns and Conventions

These are the practical patterns you'll encounter throughout the codebase. They're not arbitrary—each solves a specific problem we hit building terminal UIs.

### Pages Decide, Components Render

Components don't make decisions about layout, positioning, or what to show. They accept parameters and render. A header component doesn't decide where to position itself. A footer doesn't figure out what shortcuts to show. Pages pass that information and components render it.

This keeps components reusable. The same header works on multiple pages because it doesn't assume anything about its context. If a component needs data, its parent provides it. Data flows down, always.

### Stateless by Default

Prefer stateless components. A component that just takes parameters and returns rendered output is easier to understand, test, and reuse than one that manages internal state.

Some components need state—text inputs manage cursor position and content. Those are stateful by necessity. But even stateful components keep their state minimal. They don't sprawl into managing concerns that belong to their parent.

### Dimensions Flow Down

Pages receive width and height from their parent via `SetSize()`. They do their layout math—subtract header height, account for padding—then pass constrained dimensions to children. No component reaches up to ask "how big am I?" Dimensions propagate explicitly down the tree.

When the terminal resizes, the root TUI gets a `WindowSizeMsg`, calculates content dimensions, calls `SetSize` on the current mode, which propagates down to pages and components.

### Cursor Positioning with Markers

Manually calculating cursor X/Y coordinates is fragile—change your layout and the cursor drifts. We use markers instead. A page embeds a special marker (`CursorMarker`) in its view string where the cursor should appear. The TUI extracts this marker from the final rendered output—after all composition and padding—and calculates the position from where the marker appeared.

The marker is invisible (a null byte sequence) and gets stripped before rendering. Pages don't calculate offsets or count lines. They put a marker where they want the cursor. The TUI handles the rest.

### State Management

State lives in three places with different lifetimes:

**Control plane state** is permanent. Conversation history, quality rules, services, policies—everything that matters lives in the control plane. The CLI treats it as the source of truth. The local SQLite database is a synced replica of this state, kept current by PowerSync.

**TUI state** is ephemeral. Current page, cursor position, what's focused, loading indicators. This is UI state needed for presentation. It disappears when you close the CLI. When you restart, the local database already has current data from PowerSync, and the TUI rebuilds its UI state from scratch.

**Flow state** is accumulated during multi-step processes. FlowContext collects data as onboarding steps complete. When the flow finishes, the data gets sent to the control plane and the context is discarded. This is working memory, not persistent state.

### Wiring

Composition happens in `cmd/`. That's where implementations get wired to interfaces—config store created, keyring initialized, WorkOS client configured, GraphQL client built, PowerSync started, TUI launched with all its dependencies. Services depend on interfaces, `cmd/` provides the implementations. Nowhere else does this wiring happen.

## The MCP Server

Run `tero mcp` and the CLI becomes an MCP server, exposing Tero's knowledge to coding agents like Claude Desktop and Cursor.

The MCP server is architecturally simple. It translates MCP tool calls into queries—GraphQL for the control plane, local SQLite for catalog data—and returns structured responses the agent can reason about. No UI state, no rendering, no layout. Just protocol translation.

A tool call comes in ("get services for this organization"), a query goes out (local SQLite or GraphQL), a response comes back, the tool returns it. The MCP server doesn't cache data, doesn't manage sessions beyond authentication, and doesn't make decisions. The control plane has the intelligence. The MCP server makes it accessible through the agent's protocol.
