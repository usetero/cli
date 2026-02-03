//go:generate go run ./generate

package query

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/sqlite"
)

//go:embed schema.sql
var schema string

// Tool executes read-only SQL queries against the local catalog.
type Tool struct {
	DB sqlite.DB
}

func (t Tool) Definition() chat.Tool {
	return chat.Tool{
		Name: "query",
		Description: fmt.Sprintf(`Execute a READ-ONLY SQL query against the local SQLite catalog.

The catalog contains telemetry data synced from the Tero control plane, scoped to the user's account.

Schema:
%s

Guidelines:
- Use SELECT only (no INSERT, UPDATE, DELETE)
- Query *_cache tables for pre-computed status and metrics
- Join on id/foreign key columns (all UUIDs stored as TEXT)
- Timestamps are ISO 8601 strings
- Use LIKE for pattern matching, not regex`, schema),
		InputSchema: chat.Schema{
			Type: "object",
			Properties: map[string]chat.Property{
				"sql": {
					Type:        "string",
					Description: "The SQL query to execute",
				},
			},
			Required: []string{"sql"},
		},
	}
}

type queryInput struct {
	SQL string `json:"sql"`
}

func (t Tool) Execute(input json.RawMessage) (any, error) {
	var in queryInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Get a dedicated connection and set read-only mode
	conn, err := t.DB.Raw().Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// PRAGMA query_only prevents any writes on this connection
	if _, err := conn.ExecContext(ctx, "PRAGMA query_only = ON"); err != nil {
		return nil, err
	}

	rows, err := conn.QueryContext(ctx, in.SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, rows.Err()
}
