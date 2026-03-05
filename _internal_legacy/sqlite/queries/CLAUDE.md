# SQLite Queries

This file mirrors `AGENTS.md` for tools that still load `CLAUDE.md`.
Authoritative source: `AGENTS.md`.

This directory contains sqlc source queries. Generated Go lives in `internal/sqlite/gen`.

## Query Naming

Use explicit intent and scope in names:

1. Prefix by operation: `Get`, `List`, `Count`, `Insert`, `Update`, `Delete`, `Set`.
2. Include filter scope: `ByAccount`, `ByConversation`, etc.
3. Encode sort in the name when SQL has `ORDER BY`.

Examples:

1. `ListMessagesByConversationByCreatedAsc`
2. `ListWasteCategoryStatusesByCost`

## Ordering Rules

1. If SQL orders data, query name must indicate that order.
2. If a caller needs a different order and query uses `LIMIT`, add a new query.
3. For small result sets, re-sort in wrapper code (not in generated code).

## Null Handling Conventions

1. Use `COALESCE` for nullable text where empty string is acceptable.
2. Keep nullable floats nullable when semantics require distinguishing missing vs zero.
3. Prefer consistent parameter style across files.
