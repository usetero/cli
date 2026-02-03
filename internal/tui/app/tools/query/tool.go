//go:generate go run ./generate

package query

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/chat"
)

//go:embed schema.sql
var schema string

// Tool executes read-only SQL queries against the local catalog.
type Tool struct {
	// QueryFunc executes a query and returns results.
	// Injected by the chat model at runtime.
	QueryFunc func(sql string) ([]map[string]any, error)
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

	if t.QueryFunc == nil {
		return nil, fmt.Errorf("query function not configured")
	}

	return t.QueryFunc(in.SQL)
}
