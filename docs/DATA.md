# Data

How data flows in the CLI. Read this before diving into any interface.

## Two Data Access Patterns

The CLI has two ways to access data, depending on the interface:

**TUI and MCP** (long-lived) sync the catalog to local SQLite, then query locally:
```
PowerSync ──▶ SQLite ──▶ TUI/MCP queries
```

**CLI** (short-lived) queries the API directly:
```
CLI command ──▶ GraphQL API ──▶ response ──▶ exit
```

Why the difference? TUI/MCP are interactive—users browse, filter, search repeatedly. Local queries are fast and work offline. CLI commands are one-shot—`tero services list` runs, prints, exits. Can't wait for sync.

## Local SQLite (TUI/MCP)

When TUI or MCP starts, PowerSync syncs the catalog to local SQLite:

```
Control Plane                          CLI
┌──────────────────┐                   ┌──────────────────┐
│  Master Catalog  │ ──── PowerSync ──▶│     SQLite       │
│   (services,     │      (sync)       │  (local copy)    │
│   log events,    │                   │                  │
│   metrics...)    │                   │                  │
└──────────────────┘                   └──────────────────┘
```

From then on, all queries hit SQLite. Fast, consistent, works offline.

**What's in SQLite:**
- Catalog entities (services, log events, metrics)
- Conversations and messages (chat history)
- Context references (what entities are in a conversation)

## Direct API (CLI)

CLI commands don't sync. They call the GraphQL API directly:

```go
// In a CLI command
services, err := apiClient.Services.List(ctx)
```

This is simpler and appropriate for one-shot commands. As we add more commands (`tero services checkout`, `tero policies list`, etc.), they'll follow this pattern.

## What's Local vs Remote

| Data | TUI/MCP | CLI |
|------|---------|-----|
| Catalog (services, log events) | Local SQLite | Direct API |
| Conversations and messages | Local SQLite | N/A |
| Chat inference | Remote (Chat API) | N/A |
| Raw telemetry (actual log lines) | Remote (proxied) | Remote (proxied) |

**Key distinction:** The catalog *describes* telemetry. It doesn't *contain* it.

When you see "checkout-service has 1.2M log events," that count comes from the catalog. When you view actual log lines, those come from Datadog (or wherever) via the control plane proxy. Raw telemetry is always remote—too large to sync.

## Chat Data Flow

Chat is stateless. The Chat API doesn't remember conversations—you send the full history on every request, it streams back a response.

```
User types message
       │
       ▼
┌──────────────────┐
│  Write to SQLite │ ◀─── Immediate, before API call
└────────┬─────────┘
         │
         ▼
┌──────────────────┐     ┌──────────────────┐
│  Chat API call   │────▶│  Stream response │
│  (full history)  │     │  (text, tools)   │
└──────────────────┘     └────────┬─────────┘
                                  │
                                  ▼
                         ┌──────────────────┐
                         │  Write to SQLite │
                         └────────┬─────────┘
                                  │
                                  ▼
                         ┌──────────────────┐
                         │  Upload queue    │ ◀─── Background, for durability
                         │  (async)         │
                         └──────────────────┘
```

**Key points:**
- SQLite is written first, uploaded later. Chat works offline until you need inference.
- The Chat API is a pure function: messages in, response out. No server-side state.
- Upload queue runs in the background for durability. Can be behind—doesn't affect correctness.

## Message Storage

Messages are stored as structured content blocks:

```go
type Message struct {
    ID             MessageID
    ConversationID ConversationID
    Role           Role           // "user" or "assistant"
    Content        []Block        // The actual content
    Model          string         // e.g., "claude-3-5-sonnet"
    StopReason     string         // e.g., "end_turn" or "tool_use"
}

type Block struct {
    Type       BlockType
    Text       *TextBlock       // For text content
    Thinking   *ThinkingBlock   // For thinking content
    ToolUse    *ToolUse         // For tool calls
    ToolResult *ToolResult      // For tool results
}
```

This structure mirrors the Chat API exactly. What you send is what the model sees.

## Code Location

```
internal/sqlite/     Database interface, queries, schema
internal/powersync/  Sync engine wrapper (TUI/MCP only)
internal/api/        GraphQL client (used by CLI, also TUI/MCP for mutations)
internal/chat/       Chat client, streaming, accumulator
internal/upload/     Background upload queue
internal/domain/     Message, Block, and other types
```
