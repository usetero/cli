# MCP

Model Context Protocol server for coding agents. Not yet implemented.

## What It Will Be

MCP is a protocol for AI agents to interact with tools. Claude Code, Cursor, and similar agents connect via JSON-RPC over stdio.

```
┌─────────────┐     stdio      ┌─────────────┐
│ Claude Code │ ◀────────────▶ │  tero mcp   │
│  (agent)    │   JSON-RPC     │  (server)   │
└─────────────┘                └─────────────┘
```

The server will:
- Start PowerSync and sync the catalog (like TUI)
- Expose tools for querying the catalog
- Handle tool calls from agents

## Planned Tools

Same tools as TUI chat, plus potentially more:

- **query** — Execute read-only SQL against the catalog
- **start_journey** / **end_journey** — Guided workflows
- Additional tools as needed for agent workflows

## Data Access

MCP is long-lived like TUI. It will sync data locally via PowerSync:

```
PowerSync ──▶ SQLite ──▶ MCP tool handlers
```

This matches TUI's pattern. Agents get fast, consistent access to the catalog.

## Code Location (Planned)

```
internal/mcp/           MCP server implementation
├── server.go          JSON-RPC server
├── tools/             Tool handlers
└── ...
```

Tool definitions will likely be shared with `internal/tui/app/tools/`.
