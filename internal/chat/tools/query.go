package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

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
		Description: fmt.Sprintf(`Read-only SQL against the local SQLite catalog (synced from Tero control plane).

## Schema

%s

## Query rules

1. SELECT only — the database is read-only.
2. Prefer *_cache tables for current status and metrics.
3. Timestamps are ISO 8601 strings. Use LIKE for pattern matching.

## JOINs

JOIN on a table's id column is fast:
  JOIN log_events le ON le.id = p.log_event_id   -- PK lookup, indexed

JOIN on any other column (service_id, log_event_id, etc.) is a full table scan and will freeze the UI. Use a correlated subquery instead:

  -- WRONG: hangs for 30-60s
  LEFT JOIN log_event_statuses_cache les ON les.log_event_id = p.log_event_id

  -- RIGHT: instant
  (SELECT les.volume_per_hour FROM log_event_statuses_cache les WHERE les.log_event_id = p.log_event_id) AS volume_per_hour

Pull each column you need as a separate subquery. This applies to all tables.`, querySchema),
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

	// Reject queries with full table scans on JOINed tables — these hang for 30-60s.
	if err := t.checkQueryPlan(ctx, in.SQL); err != nil {
		return tools.QueryResult{}, err
	}

	// Use the read pool — every connection has query_only = ON enforced by the driver
	rows, err := t.db.ReadRaw().QueryContext(ctx, in.SQL)
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

	capped, dropped := capResults(results)
	return tools.QueryResult{Rows: capped, RowsDropped: dropped}, nil
}

const (
	maxFieldBytes  = 4096   // truncate any string value longer than this
	maxResultBytes = 102400 // drop trailing rows if total JSON exceeds this
)

// capResults truncates large string values and drops trailing rows to keep
// the serialized result within size limits. Returns the capped rows and the
// number of rows that were dropped.
func capResults(rows []map[string]any) ([]map[string]any, int) {
	for _, row := range rows {
		for k, v := range row {
			s, ok := v.(string)
			if ok && len(s) > maxFieldBytes {
				row[k] = s[:maxFieldBytes] + "…(truncated)"
			}
		}
	}

	b, err := json.Marshal(rows)
	if err != nil || len(b) <= maxResultBytes {
		return rows, 0
	}

	// Binary search for the max number of rows that fit.
	total := len(rows)
	lo, hi := 0, total
	for lo < hi {
		mid := (lo + hi + 1) / 2
		b, _ = json.Marshal(rows[:mid])
		if len(b) <= maxResultBytes {
			lo = mid
		} else {
			hi = mid - 1
		}
	}

	return rows[:lo], total - lo
}

// checkQueryPlan runs EXPLAIN QUERY PLAN and rejects queries that would do
// a full table scan on a JOINed table (SCAN ... JOIN). These take 30-60s
// due to a SQLite limitation with expression indexes through views.
func (t *QueryTool) checkQueryPlan(ctx context.Context, sql string) error {
	rows, err := t.db.ReadRaw().QueryContext(ctx, "EXPLAIN QUERY PLAN "+sql)
	if err != nil {
		return nil // let the actual query surface the error
	}
	defer rows.Close()

	var scans []string
	for rows.Next() {
		var id, parent, notUsed int
		var detail string
		if err := rows.Scan(&id, &parent, &notUsed, &detail); err != nil {
			return nil
		}
		if strings.Contains(detail, "SCAN") && strings.Contains(detail, "JOIN") {
			scans = append(scans, detail)
		}
	}

	if len(scans) == 0 {
		return nil
	}

	return fmt.Errorf(
		"query rejected: full table scan detected on a JOINed table, which would hang for 30-60 seconds.\n\n"+
			"Problem: %s\n\n"+
			"Fix: replace the JOIN with a correlated subquery. "+
			"JOIN on a table's id column is fast (e.g., JOIN log_events le ON le.id = p.log_event_id). "+
			"For any other column, use: (SELECT col FROM table WHERE foreign_key = outer.value) AS alias",
		strings.Join(scans, "; "),
	)
}
