# Data Flow

This repo uses two data flows.

## 1) Direct Query Flow (CLI)

1. User runs a command.
2. Command calls `internal/api`.
3. Control plane responds.
4. Output is rendered directly.

Use this for stateless command execution.

## 2) Sync + Local Read Flow (TUI)

1. PowerSync syncs remote data into local SQLite.
2. TUI reads from local SQLite for responsive rendering.
3. Mutations go to control plane; sync converges local state.

Use this for interactive views and offline-tolerant UX.

## Chat Data Notes

1. Message history is persisted in SQLite.
2. Streaming assistant events are reduced before UI application.
3. Turn scoping is mandatory for streamed content and tool completions.
4. User-cancelled partial assistant output is not persisted as a committed assistant turn.

