# TUI Architecture

- [1. Introduction](#1-introduction)
- [2. The Elm Architecture](#2-the-elm-architecture)
- [3. Structure](#3-structure)
- [4. State Management](#4-state-management)
- [5. Component Philosophy](#5-component-philosophy)
- [6. Layout and Rendering](#6-layout-and-rendering)
- [7. Content Blocks](#7-content-blocks)

---

## 1. Introduction

The TUI (Terminal User Interface) is the primary way users interact with Tero. When you run `tero`, you get a rich, interactive terminal experience—conversational, stateful, and beautiful. This isn't a traditional command-line tool where you type commands and get text output. It's a full application that happens to run in your terminal.

The TUI is built on [Bubbletea](https://github.com/charmbracelet/bubbletea), a Go framework based on The Elm Architecture. This choice isn't arbitrary—building complex, stateful terminal interfaces is hard. Bubbletea makes it manageable by providing a clear pattern for state management, predictable updates, and composable components.

This document explains how the TUI is architected. It covers the foundational patterns, the structure, and the design decisions that keep the codebase maintainable as complexity grows. If you're working on TUI features—adding pages, building components, rendering content—this is your guide to understanding how it all fits together.

## 2. The Elm Architecture

Bubbletea is built on [The Elm Architecture](https://guide.elm-lang.org/architecture/), a pattern for building UIs that makes state management predictable. The pattern is simple: your application is a model (all your state), an update function (how state changes), and a view function (how to render state). Messages flow in—keypresses, API responses, timer ticks—and trigger updates. The model changes, the view re-renders, the cycle continues.

This matters for us because the TUI has a lot of state to manage. Which page are you on? What's the conversation history? Where's the cursor? What's focused? Without a pattern, this gets messy fast. The Elm Architecture keeps it organized—all state lives in the model, all changes happen in update functions, and views are always pure functions of state.

The pattern also makes the code composable. Pages are models with update and view functions. Components are models with update and view functions. The same pattern repeats at every level, which means once you understand it in one place, you understand it everywhere.

If you're new to this pattern, the Elm Architecture guide linked above explains it well. For our purposes, just know: state lives in structs, changes happen in update functions, and views are pure functions that take state and return strings. Everything else follows from that.

One implementation detail worth understanding: Elm enforces immutability—every update returns a new value. Bubbletea follows this pattern with value receivers, where `Update()` returns the new state. This works well for top-level models (pages, modes) where you want state transitions to be explicit. A page completing? Return the new page. Clear transition.

But Go isn't Elm. We don't have persistent data structures or compiler-enforced immutability. For leaf components (footer, header, text input), the value receiver pattern creates a footgun: forgetting to capture the return value from `Update()` silently loses state changes. So our components use pointer receivers instead—standard Go for methods that mutate. This is what Crush (Charm's production app) does. They learned this building a real application.

The mental model: pages and modes are Elm-style (explicit state transitions via value receivers), components are Go-style (direct mutation via pointer receivers). This matches both the framework's intent and what actually works in practice.

## 3. Structure

The TUI is organized hierarchically. At the top is the root TUI model that manages global concerns—current page, session state, dimensions. Below that are pages—onboarding, chat, and eventually others. Pages manage their own UI and compose together reusable components like headers, footers, and input fields.

Think of it as layers of responsibility. The root TUI knows about page routing and global state. Pages know about their specific UI flow and what components they need. Components know how to render specific pieces of UI. Each layer has a clear job.

### The Root TUI

Lives in `internal/tui/tui.go`. This is the entry point when you run `tero`. It initializes the application with services and dependencies, creates a Flow starting with the auth page, and manages the global footer that appears on every screen.

The root TUI's job is coordination. It doesn't render much itself—it delegates to the Flow, which manages page transitions. The TUI composes the current page's output with a footer, handles window resizing, and manages the progress bar for supported terminals. When the Flow completes (onboarding finishes), the TUI transitions to the chat page.

### Flow and Pages

Page transitions follow a simple chain pattern. Each page knows what comes next. When a page completes its work, it returns its successor via the `Next()` method. The Flow orchestrator handles the mechanics—checking if the current page is complete, calling Next() to get the successor, initializing the new page with proper dimensions.

This is a linked list where each page is a node that knows its successor. The auth page returns the onboarding page. The role selection step returns the organization selection step. The final onboarding step returns `nil`, signaling completion. No arrays, no index tracking, just pages forming a chain.

The Flow lives in `internal/tui/pages/flow.go`. It's simple—about 80 lines. It holds the current page and a context bag. On each update, it forwards messages to the current page. After updating, it checks `IsComplete()`. If true, it calls `Next()` to get the next page, sets dimensions on it, initializes it, and continues. The pattern is consistent whether you're in a multi-step onboarding flow or a top-level page transition.

Pages implement a common interface with six methods: `Init()`, `Update()`, `View()`, `SetSize()`, `IsComplete()`, and `Next()`. Some pages are complete experiences—the chat page. Others are wrappers that add chrome and delegate to an inner flow—the onboarding page wraps a flow of steps and adds a header. The pattern works at multiple levels because it's the same interface everywhere.

### Components

Components are reusable pieces that pages compose together. Headers show the Tero logo and context. Footers show shortcuts and version info. Text inputs handle user input with cursors and validation. Each component is focused on doing one thing well.

Components come in two flavors: stateful (like text inputs that manage cursor position) and stateless (like headers that just render what they're given). We prefer stateless when possible—less state means less complexity.

## 4. State Management

State in the TUI lives in three places, each with a different lifetime and purpose. Understanding these separations is key to working with the codebase effectively.

### Control Plane State (Permanent)

The control plane owns all the important data. Conversation history, quality rules, your services, user information—everything that matters permanently lives there. When you send a message, the control plane adds it to the conversation. When you close the CLI and reopen it, the control plane still has everything.

The TUI treats the control plane as the source of truth. It never tries to be authoritative about data. It asks the control plane for information, caches it locally for fast rendering, but knows the control plane is always right.

### TUI State (Ephemeral)

The TUI model holds state needed for presentation. Current page, cursor position, what's focused, loading indicators, optimistic updates—this is UI state that helps make the interface responsive and smooth.

This state disappears when you close the CLI. It's ephemeral by design. The TUI doesn't try to persist anything. When you restart, it fetches fresh state from the control plane and rebuilds the UI.

### Flow State (Accumulated)

As pages chain together in a Flow, they need to share data. The email you enter in auth needs to reach the organization creation step. The organization you select needs to reach account creation. This accumulated state lives in FlowContext, a mutable bag passed to every page's `Next()` method.

FlowContext has two kinds of state. First, services—the GraphQL client, preferences storage, logger. These are initialized once at the start and don't change. Every page gets access to the same service instances. Second, accumulated data—email, selected organization ID, chosen role. These fields start empty and get filled as the flow progresses.

When a page completes and calls `Next()`, it receives the FlowContext. It updates the relevant fields with its own results, then passes the context to the next page's constructor. The pattern is simple: each page adds its data to the shared context and passes it forward. By the time you reach the end of onboarding, the context has everything—email, role, organization, account.

During onboarding, you enter an email, select a role, pick an organization, create an account. Each step stores its result in the context. When onboarding finishes, all that accumulated data gets sent to the control plane in one or more API calls, then the local context is discarded.

This is different from TUI state (which is about presentation) and control plane state (which is permanent). Flow state is working memory for multi-step processes. It exists while you're in the flow, disappears when the flow completes. It's pragmatic state management—no clever abstractions, no complex state machines. Just a struct with fields that pages read from and write to as they complete their work.

### The Pattern

When you need data, ask the control plane via GraphQL. When you get a response, store it in the TUI model for rendering. When the user takes action, send it to the control plane and update the local model optimistically. When the control plane confirms, update the model with the real response.

For multi-step flows, accumulate data in FlowContext as you progress. When the flow completes, send everything to the control plane. The context then gets discarded—it served its purpose.

This keeps things simple. The TUI doesn't manage databases or worry about persistence. It's a view into the control plane's state, with enough local caching to feel fast and responsive.

## 5. Component Philosophy

The way we think about components shapes how maintainable the TUI stays as it grows. The core principle is simple: components render, pages decide. Components don't make decisions about layout, positioning, or what to show—they accept parameters and render beautifully.

### Pages Decide, Components Render

Pages know the context. They know what data to show, where to position things, how much space is available. Components just know how to render specific pieces of UI when given the right inputs.

A header component doesn't decide where to position itself or how wide it should be. The page tells it "render at this width" and the header does exactly that. A footer doesn't figure out what shortcuts to show—the page passes that information and the footer renders it.

This separation keeps components reusable. The same header component works on multiple pages because it doesn't assume anything about its context. The same footer works everywhere because pages tell it what to display.

### Stateless by Default

Prefer stateless components. A component that just takes parameters and returns rendered output is easier to understand, test, and reuse than one that manages internal state.

The header component? Stateless. You pass it a width, it renders. The footer? Stateless. You pass it text, it renders. These components are pure functions—same inputs always produce the same output.

Some components need state. Text inputs manage cursor position and content. Those are stateful by necessity. But even stateful components keep their state minimal and well-defined. They don't sprawl into managing concerns that belong to their parent.

### Props Down

Data flows down. Pages pass data to components through parameters. Components never reach up to grab data from their parents or make GraphQL calls themselves. If a component needs data, its parent provides it.

This makes the data flow explicit and traceable. You can look at a page's view function and see exactly what data goes to each component. No hidden dependencies, no surprise side effects.

## 6. Layout and Rendering

Layout in terminal UIs is different from web UIs. You don't have CSS. You don't have flexbox. You have characters, line widths, and manual positioning. The patterns we've developed keep this manageable without becoming brittle.

### Nested Pages and Delegation

Some pages are wrappers that add visual chrome and delegate to inner flows. The onboarding page is a good example. It wraps an inner Flow of steps (role selection, organization creation, account setup) and adds a header with the Tero logo. The outer page handles layout—where the header goes, how content is positioned. The inner flow handles orchestration—which step you're on, when to transition.

This nesting works because everything implements the same Page interface. The onboarding page receives messages, delegates them to its inner flow, and returns itself. When asked for a view, it renders its header, gets the current step's view from the flow, positions them together, and returns the composed result. The pattern is clean: outer pages add chrome and delegate behavior.

You can read this in `onboarding.go`. The Update method forwards messages to the flow. The View method renders the header, asks the flow for its view, calculates spacing, and composes them together. Methods like `IsComplete()`, `HelpText()`, and `IsBusy()` all delegate to the inner flow. The onboarding page is thin—it adds visual structure but doesn't duplicate logic.

This pattern scales. If chat needs a sidebar, it wraps a flow and adds sidebar chrome. If settings needs a navigation menu, it wraps content pages and adds menu chrome. The outer page handles "where things go," the inner pages handle "what to show." Separation of concerns through delegation.

### Dimensions Flow Down

Pages receive width and height from their parent via `SetSize()`. They use these dimensions to lay out their content, then pass appropriate dimensions to any child components or flows they contain.

This is straightforward propagation—parents tell children their constraints, children render within those bounds. A page that shows a header subtracts the header's height from the total before telling its content area how much space it has. A page that adds padding reduces width and height before passing dimensions to children.

The pattern is always the same: receive dimensions, do your layout math, pass constrained dimensions to children. No components reaching up to ask "how big am I?" or making assumptions. Dimensions flow explicitly down the tree. When the terminal resizes, the root TUI gets a WindowSizeMsg, calculates content dimensions, calls SetSize on the current page, which propagates down to components and flows.

### Cursor Positioning with Markers

Terminal applications need to tell the terminal where to draw the cursor. Traditionally, this meant manually calculating X/Y coordinates—count the lines above your input, measure the characters to the left, add offsets for padding. It works, but it's fragile. Change your layout and suddenly your cursor is off by one.

We use markers instead. When a page wants to show a cursor, it embeds a special marker (`pages.CursorMarker`) in its view string at exactly the position where the cursor should appear. The TUI extracts this marker from the final rendered output—after all composition, padding, and layout is complete—and calculates the cursor position from where the marker appeared in the final string.

This makes cursor positioning robust. Pages don't calculate offsets or count lines. They just put a marker where they want the cursor. The TUI handles the rest. Change your layout, add padding, compose things differently—the cursor stays in the right place because it's defined relative to the actual rendered output, not manual calculations.

The marker is invisible (it's a null byte sequence) and gets stripped before rendering. Users never see it. It's purely a positioning mechanism that makes our code simpler and more maintainable.

## 7. Content Blocks

The control plane sends responses as content blocks—structured pieces of data that represent different types of content. The TUI's job is to render these beautifully in the terminal.

A content block might be text (a natural language response), a chart (time series data showing trends), a table (services and their metrics), log samples (with syntax highlighting), or action buttons (things the user can do). The control plane decides what blocks to send based on the conversation. The TUI decides how to render each type.

This separation is powerful. The control plane doesn't need to know about terminal capabilities, color schemes, or layout constraints. It just sends structured data—"here's a time series with these data points" or "here's a table with these columns and rows." The TUI handles all the presentation details.

Adding new content block types follows a pattern. The control plane defines the new block structure in its GraphQL schema. The CLI regenerates its GraphQL client to pick up the new types. Then we add a renderer for that block type in the TUI. Old CLI versions that don't know about the new block type simply ignore it—graceful degradation built in.

The rendering implementations live in the TUI codebase, each focused on making its specific content type look good in a terminal. Charts become ASCII visualizations. Tables get borders and alignment. Logs get syntax highlighting. The goal is always the same: take the control plane's structured data and make it beautiful, readable, and actionable in the terminal.
