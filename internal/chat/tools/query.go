//go:generate go run ./generate

package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/sqlite"
)

//go:embed query_schema.sql
var querySchema string

// QueryTool executes read-only SQL queries against the local catalog.
type QueryTool struct {
	db sqlite.DB
}

// NewQueryTool creates a new query tool.
func NewQueryTool(db sqlite.DB) *QueryTool {
	return &QueryTool{db: db}
}

// Name returns the tool name.
func (t *QueryTool) Name() string {
	return "query"
}

// Definition returns the tool definition for the chat API.
func (t *QueryTool) Definition() chat.Tool {
	return chat.Tool{
		Name: t.Name(),
		Description: fmt.Sprintf(`Execute a READ-ONLY SQL query against the local SQLite catalog.

The catalog contains telemetry data synced from the Tero control plane, scoped to the user's account.

Schema:
%s

Guidelines:
- Use SELECT only (no INSERT, UPDATE, DELETE)
- Query *_cache tables for pre-computed status and metrics
- Join on id/foreign key columns (all UUIDs stored as TEXT)
- Timestamps are ISO 8601 strings
- Use LIKE for pattern matching, not regex`, querySchema),
		InputSchema: chat.NewObjectSchema(
			map[string]chat.Property{
				"sql": {
					Type:        "string",
					Description: "The SQL query to execute",
				},
				"status": {
					Type:        "string",
					Description: "Message shown while the query runs (e.g., 'Checking service status', 'Looking for errors')",
				},
				"result": {
					Type:        "string",
					Description: "Message shown when complete. Use {count} for row count if relevant (e.g., 'Found {count} services'). Omit {count} for aggregates (e.g., 'Calculated')",
				},
			},
			[]string{"sql", "status", "result"},
		),
	}
}

// Execute runs the query and returns a typed result.
func (t *QueryTool) Execute(input json.RawMessage) (tools.QueryResult, error) {
	var in tools.QueryInput
	if err := json.Unmarshal(input, &in); err != nil {
		return tools.QueryResult{}, err
	}

	ctx := context.Background()

	// Get a dedicated connection and set read-only mode
	conn, err := t.db.Raw().Conn(ctx)
	if err != nil {
		return tools.QueryResult{}, err
	}
	defer func() {
		// Reset query_only before returning connection to pool
		_, _ = conn.ExecContext(ctx, "PRAGMA query_only = OFF")
		conn.Close()
	}()

	// PRAGMA query_only prevents any writes on this connection
	if _, err := conn.ExecContext(ctx, "PRAGMA query_only = ON"); err != nil {
		return tools.QueryResult{}, err
	}

	rows, err := conn.QueryContext(ctx, in.SQL)
	if err != nil {
		return tools.QueryResult{}, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return tools.QueryResult{}, err
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return tools.QueryResult{}, err
		}

		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return tools.QueryResult{}, err
	}
	return tools.QueryResult{Rows: results}, nil
}
