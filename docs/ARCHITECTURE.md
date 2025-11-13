# CLI Architecture

- [1. Introduction](#1-introduction)
- [2. What the CLI Does](#2-what-the-cli-does)
- [3. How It Works (End to End)](#3-how-it-works-end-to-end)
- [4. The App Layer](#4-the-app-layer)
- [5. Commands and Modes](#5-commands-and-modes)
- [6. The TUI](#6-the-tui)
- [7. The MCP Server](#7-the-mcp-server)

---

## 1. Introduction

Tero is a control plane for observability. It sits on top of your existing tools (Datadog, Splunk, CloudWatch, etc.), understands what your telemetry means semantically, and helps you improve it—identify waste, take action, reduce cost.

The CLI is a tool for interacting with Tero. It provides multiple interfaces depending on how you want to work:
- An interactive TUI (Terminal User Interface) for conversational exploration
- An MCP server that exposes Tero's knowledge to coding agents
- Traditional commands for scripting and automation (future)

This document explains the CLI's architecture: how it's structured, how it communicates with the control plane, and how the pieces fit together. It complements [TUI.md](TUI.md), which dives deep into the Terminal User Interface, and [DESIGN.md](DESIGN.md), which explains the UX principles that guide our decisions.

## 2. What the CLI Does

The CLI is the **presentation layer** for Tero. The control plane does the hard work—analyzing telemetry, building semantic catalogs, classifying quality, identifying waste. The CLI's job is to make that intelligence accessible and actionable through beautiful, intuitive interfaces.

Think of it this way: the control plane is the brain. The CLI is how you interact with it. The control plane holds all the knowledge, runs all the analysis, and stores all the data permanently. The CLI presents that knowledge, lets you explore it conversationally, and helps you take action based on what you learn.

This separation runs deep. The CLI never implements intelligence. It doesn't analyze log patterns, calculate waste percentages, or make decisions about data quality. When you're working on the CLI and you're tempted to add logic that feels smart—stop. That logic belongs in the control plane, exposed via GraphQL, and consumed by the CLI for presentation.

### Presentation vs Intelligence

Here's the key principle: **the control plane sends data, the CLI decides how to present it.**

The control plane says: "Here's time series data showing waste over time."

The CLI decides: "I'll render that as a line chart in the terminal."

This separation means the CLI can evolve independently. Add new chart types? Improve layouts? Enhance interactions? No control plane changes required. The control plane sends structured data—what it means, what it represents. The CLI handles all presentation concerns—how it looks, how users interact with it, how it adapts to different contexts.

The CLI is stateful while running—it caches conversations, maintains UI state, remembers where you are. But when you close it, that state disappears. The control plane is the permanent source of truth. Everything that matters—conversation history, quality rules, your services and their data—lives in the control plane. The CLI is just a window into that truth.

## 3. How It Works (End to End)

At its core, the CLI is a client to the control plane's GraphQL API. Whether you're using the TUI, the MCP server, or traditional commands, the pattern is the same: authenticate, communicate via GraphQL, present results.

### Authentication

The CLI uses WorkOS for authentication. First-time users go through signup—email collection, organization creation or discovery, SSO or email/password flow. Returning users have stored credentials that the CLI validates or refreshes as needed.

Once authenticated, the CLI has a token it includes with every GraphQL request. The control plane uses this to identify the user, their organization, and their permissions.

### Communication

All communication happens through GraphQL. The CLI sends queries (read data) and mutations (write data) to the control plane. The control plane processes these requests with full context—who you are, what you've done before, what your organization looks like—and returns structured responses.

The CLI doesn't manage state beyond what's needed for presentation. Session history? The control plane stores it. Quality rules? Control plane. Your services and their data? Control plane. The CLI caches some of this locally while running for fast rendering, but the control plane is always the source of truth.

### Content Blocks

Responses from the control plane come as content blocks—structured pieces of data that the CLI knows how to render. Text blocks, chart data, table data, log samples, actions the user can take. The control plane decides what content is relevant based on the conversation. The CLI decides how to present it based on the interface (TUI renders charts visually, MCP returns structured data to coding agents).

This separation means new content types can be added over time. The control plane starts sending a new block type, old CLI versions ignore what they don't understand, new CLI versions render it beautifully.

## 4. The App Layer

Between the interfaces (TUI, MCP server) and the control plane sits the app layer—the business logic that orchestrates authentication, manages user preferences, and translates control plane responses into domain models the interfaces can work with.

The app layer is structured in three tiers: services define domain operations, interfaces abstract dependencies, and implementations provide the concrete machinery. This separation makes the code testable, maintainable, and adaptable as requirements evolve.

### Services

Services own domain concepts and orchestrate operations. `PreferencesService` knows about email addresses and organization IDs. `AuthService` knows about device authorization flows and token refresh. `ServiceService` knows how to list services discovered from Datadog, check discovery status, and retrieve counts.

Services don't know about YAML files, OS keychains, or GraphQL queries. They work with interfaces—`Store` for preferences, `SecureStorage` for tokens, `OAuthProvider` for authentication, `APIClient` for the control plane. This keeps them focused on domain logic without coupling to implementation details.

Method names follow consistent patterns: `ListByX` for collections, `GetByX` for single items, `CountByX` for totals. Services are scoped by account type where relevant—`ListByDatadogAccount` makes it clear we're querying Datadog-discovered services. When we add Splunk support, we'll add `ListBySplunkAccount` alongside it.

### Interfaces

Interfaces are defined by their consumers, not their providers. `PreferencesService` needs generic key-value storage, so it depends on `Store`—a simple interface with `Get`, `Set`, `GetBool`, `GetList`, and `Save`. It doesn't care whether the implementation uses YAML, JSON, or a database. It just needs somewhere to put key-value pairs.

This consumer-driven approach keeps interfaces small and focused. `SecureStorage` is just `Get`, `Set`, and `Delete`—generic operations that work for any secure storage mechanism. `OAuthProvider` defines the device authorization flow without mentioning WorkOS specifically. New implementations can plug in without changing the services that depend on them.

Interfaces live in the app package alongside services. This makes dependencies explicit and keeps the architecture navigable. When you read a service, you see exactly what it needs. When you implement an interface, you know exactly what contract you're fulfilling.

### Implementations

Concrete implementations handle the messy details. `config.Config` implements `Store` with YAML files. `keyring.Keyring` implements `SecureStorage` with OS keychains—Keychain on macOS, Credential Manager on Windows, Secret Service on Linux. `workos.Client` implements `OAuthProvider` with WorkOS API calls. The generated GraphQL client implements `APIClient`.

Implementations can be swapped without touching services. Want to use JSON instead of YAML for config? Implement `Store` differently. Need to support a different OAuth provider? Implement `OAuthProvider`. The services don't change—they depend on interfaces, not concrete types.

This tier handles all platform-specific concerns, external API integration, file formats, and persistence strategies. Services stay clean, focused on domain logic. Interfaces stay simple, defining only what's needed. Implementations deal with reality.

### Testing

The app layer includes `apptest`—a package of mock implementations for every interface. These mocks use function fields, making them trivial to configure in tests. Need to test what happens when authentication fails? Set `MockOAuthProvider.PollAuthenticationFunc` to return an error. Want to verify a service saves preferences correctly? Check what `MockStore.SetFunc` was called with.

This pattern makes tests fast, focused, and deterministic. No file I/O, no network calls, no OS dependencies. Just pure logic tests with complete control over dependencies.

## 5. Commands and Modes

The CLI provides different ways to interact with Tero depending on what you're trying to do. The architecture supports multiple modes through a single codebase, all sharing the same foundation: GraphQL communication, authentication, and content rendering.

### Interactive Chat

Run `tero` and you get the TUI—an interactive, conversational interface built on Bubbletea. This is the primary way most users interact with Tero. You ask questions, explore your data, and take action through natural conversation. The TUI maintains UI state (which page you're on, cursor position, what's focused) but all the data comes from and lives in the control plane.

The TUI architecture is substantial enough that it has its own documentation. See [TUI.md](TUI.md) for the deep dive on how it works—the Elm Architecture pattern, component design, layout management, and rendering strategies.

### MCP Server

Run `tero mcp` and the CLI becomes an MCP server, exposing Tero's knowledge to coding agents like Claude Desktop and Cursor. Instead of rendering content visually, the MCP server returns structured data that agents can reason about. The same GraphQL queries power both modes—the MCP server just presents results differently.

The MCP server is a thin layer. It translates MCP tool calls into GraphQL queries, sends them to the control plane, and returns the results. No intelligence, no caching, no state management. Just protocol translation.

### Traditional Commands (Future)

Eventually the CLI will support traditional command patterns for scripting and automation—`tero status`, `tero export`, `tero config`. These will use the same GraphQL client, the same authentication, but return plain text output suitable for piping and parsing. Different interface, same foundation.

## 6. The TUI

The TUI is where most of the CLI's complexity lives. It's built on Bubbletea, a Go framework based on The Elm Architecture. This architecture pattern—Model-Update-View—makes building complex, stateful terminal interfaces manageable.

The model holds all state: which page you're on, conversation history, cursor position, what's focused. Updates are the only way to change state—a keypress arrives as a message, the update function processes it, returns a new model. The view function takes the current model and renders it to the terminal. This cycle repeats continuously, making state changes predictable and debuggable.

The TUI is structured hierarchically. At the root is the main TUI model that manages global concerns—authentication state, current page, session ID. Pages manage their own UI—the onboarding page handles email collection, the chat page handles conversation display and input. Components are reusable pieces—headers, footers, input fields—that pages compose together.

This architecture has evolved specific patterns for challenges unique to terminal UIs: layout and padding (parents calculate, children accept), cursor positioning (offsets propagate down the tree), content rendering (different block types for text, charts, tables, logs). These patterns keep the code maintainable as complexity grows.

The details are substantial. For the deep dive on TUI architecture, component patterns, and development workflows, see [TUI.md](TUI.md).

## 7. The MCP Server

The MCP server is architecturally much simpler than the TUI. It's a thin translation layer between the Model Context Protocol and Tero's GraphQL API.

Run `tero mcp` and the CLI becomes an MCP server for coding agents—Claude Desktop, Cursor, and other tools that support MCP. This brings Tero's telemetry intelligence directly into developers' coding workflows. A developer working on code can ask their agent about log events, quality scores, waste patterns—and the agent has access to Tero's knowledge to answer.

The agent makes tool calls—"get services for this organization," "analyze this log statement," "what's the quality score for checkout-api." The MCP server translates these into GraphQL queries, sends them to the control plane, and returns structured responses the agent can work with.

There's no UI state, no rendering complexity, no layout management. Just protocol translation. A tool call comes in, a GraphQL query goes out, a response comes back, the tool returns it. The MCP server doesn't cache data, doesn't manage sessions beyond what's needed for authentication, and doesn't make decisions. It's a pure proxy.

This simplicity is intentional. The control plane has all the intelligence. The MCP server just makes that intelligence accessible to coding agents through their protocol.
