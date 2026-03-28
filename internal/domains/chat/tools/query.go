package tools

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	queryTimeout   = 3 * time.Second
	queryRowCap    = 500
	queryResultCap = 128 * 1024
	queryFieldCap  = 4096
)

//go:embed query_schema.sql
var querySchema string

type QueryTool struct {
	db interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}
}

type QueryInput struct {
	SQL string `json:"sql"`
}

type QueryResult struct {
	Rows        []map[string]any `json:"rows"`
	RowsDropped int              `json:"rows_dropped,omitempty"`
}

func NewQueryTool(db interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}) *QueryTool {
	if db == nil {
		panic("query tool requires db")
	}
	return &QueryTool{db: db}
}

func (t *QueryTool) Definition() Definition {
	return Definition{
		Name: QueryToolName,
		Description: fmt.Sprintf(`Run read-only SQL against local SQLite state.

Schema:

%s

Rules:
1. SELECT or WITH only.
2. Query local SQLite projections, not control-plane concepts.
3. Prefer cache/status tables for current metrics and derived state.
4. Timestamps are ISO 8601 strings.
`, strings.TrimSpace(querySchema)),
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"sql": map[string]any{
					"type":        "string",
					"description": "SELECT query to run",
				},
			},
			"required": []string{"sql"},
		},
	}
}

func (t *QueryTool) Run(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var in QueryInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("parse query input: %w", err)
	}
	sqlText := strings.TrimSpace(in.SQL)
	if sqlText == "" {
		return nil, fmt.Errorf("sql is required")
	}
	if !isReadOnlySelect(sqlText) {
		return nil, fmt.Errorf("only read-only SELECT queries are allowed")
	}

	qctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := t.db.QueryContext(qctx, sqlText)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := QueryResult{Rows: make([]map[string]any, 0, 64)}
	totalBytes := 2 // []
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
			row[col] = normalizeQueryValue(values[i])
		}

		rowBytes, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}
		nextBytes := totalBytes + len(rowBytes)
		if len(result.Rows) > 0 {
			nextBytes++
		}
		if len(result.Rows) < queryRowCap && nextBytes <= queryResultCap {
			result.Rows = append(result.Rows, row)
			totalBytes = nextBytes
		} else {
			result.RowsDropped++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func isReadOnlySelect(sqlText string) bool {
	normalized := strings.TrimSpace(strings.ToLower(sqlText))
	return strings.HasPrefix(normalized, "select ") || strings.HasPrefix(normalized, "with ")
}

func normalizeQueryValue(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case []byte:
		s := string(x)
		if len(s) > queryFieldCap {
			return s[:queryFieldCap] + "...(truncated)"
		}
		return s
	case string:
		if len(x) > queryFieldCap {
			return x[:queryFieldCap] + "...(truncated)"
		}
		return x
	default:
		return x
	}
}
