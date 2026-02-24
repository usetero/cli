package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/sqlite"
)

// ShowTool resolves entity IDs and fetches card data from the local catalog.
type ShowTool struct {
	db        sqlite.DB
	resolvers map[tools.EntityType]entityResolver
}

// entityResolver fetches card data for a specific entity type.
type entityResolver struct {
	fetch func(ctx context.Context, id string) (tools.ShowResult, error)
}

// NewShowTool creates a new show tool.
func NewShowTool(db sqlite.DB) *ShowTool {
	policyStatuses := db.LogEventPolicyStatuses()
	return &ShowTool{
		db: db,
		resolvers: map[tools.EntityType]entityResolver{
			tools.EntityPolicy: {fetch: func(ctx context.Context, id string) (tools.ShowResult, error) {
				return fetchPolicyCard(ctx, policyStatuses, id)
			}},
		},
	}
}

// Name returns the tool name.
func (t *ShowTool) Name() string { return "show" }

// Definition returns the tool definition for the chat API.
func (t *ShowTool) Definition() chat.Tool {
	return chat.Tool{
		Name: t.Name(),
		Description: `Show a single entity as a rich, formatted card rendered inline in the conversation. The user sees the card directly — do NOT repeat or summarize the card contents in your response. Just reference it conversationally.

## How it works

The card fetches and displays all relevant data automatically from the entity ID. You provide the entity type and either an ID or a lookup query.

## Input

Provide either:
- "id" — the entity's UUID (if you already have it from a previous query result)
- "sql" — a SELECT that returns exactly one row with an "id" column (when you need to look up the entity)

## Entity: policy

Look up policy IDs from log_event_policy_statuses_cache:
  SELECT policy_id AS id FROM log_event_policy_statuses_cache WHERE service_name = '...' AND log_event_name = '...' AND category = '...'

The card displays: category, service, log event, action, status, severity, volume, throughput, and estimated savings.

## When to use show vs query

- Use "show" when presenting a specific entity to the user — it renders a styled card.
- Use "query" for data exploration, aggregations, comparisons, or when you need raw tabular results.`,
		InputSchema: chat.NewObjectSchema(
			map[string]chat.Property{
				"entity": {
					Type:        "string",
					Enum:        []string{"policy"},
					Description: "The entity type to show.",
				},
				"id": {
					Type:        "string",
					Description: "UUID of the entity. Provide this OR sql, not both.",
				},
				"sql": {
					Type:        "string",
					Description: "SQL query returning exactly one row with an 'id' column. Used when you need to look up the entity. Provide this OR id, not both.",
				},
				"title": {
					Type:        "string",
					Description: "Optional title shown above the card.",
				},
			},
			[]string{"entity"},
		),
	}
}

// Execute resolves the entity ID and fetches card data.
func (t *ShowTool) Execute(input json.RawMessage) (tools.ShowResult, error) {
	var in tools.ShowInput
	if err := json.Unmarshal(input, &in); err != nil {
		return tools.ShowResult{}, err
	}

	resolver, ok := t.resolvers[in.Entity]
	if !ok {
		return tools.ShowResult{}, fmt.Errorf("unsupported entity type: %q", in.Entity)
	}

	id := in.ID
	if id == "" && in.SQL != "" {
		resolved, err := t.resolveIDFromSQL(in.SQL)
		if err != nil {
			return tools.ShowResult{}, err
		}
		id = resolved
	}
	if id == "" {
		return tools.ShowResult{}, fmt.Errorf("either id or sql must be provided")
	}

	ctx := context.Background()
	result, err := resolver.fetch(ctx, id)
	if err != nil {
		return tools.ShowResult{}, err
	}
	result.Title = in.Title
	return result, nil
}

// resolveIDFromSQL runs a SQL query and extracts the id column from the single result row.
func (t *ShowTool) resolveIDFromSQL(query string) (string, error) {
	ctx := context.Background()

	if err := t.checkQueryPlan(ctx, query); err != nil {
		return "", err
	}

	rows, err := t.db.ReadRaw().QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("sql lookup failed: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	idIdx := -1
	for i, col := range cols {
		if col == "id" {
			idIdx = i
			break
		}
	}
	if idIdx == -1 {
		return "", fmt.Errorf("sql query must return an 'id' column; got columns: %v", cols)
	}

	if !rows.Next() {
		return "", fmt.Errorf("sql query returned no rows")
	}

	values := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return "", err
	}

	if rows.Next() {
		return "", fmt.Errorf("sql query returned more than one row; must return exactly one")
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	id, ok := values[idIdx].(string)
	if !ok {
		return "", fmt.Errorf("id column is not a string: %v", values[idIdx])
	}
	return id, nil
}

// checkQueryPlan rejects queries with full table scans on JOINed tables.
func (t *ShowTool) checkQueryPlan(ctx context.Context, query string) error {
	rows, err := t.db.ReadRaw().QueryContext(ctx, "EXPLAIN QUERY PLAN "+query)
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

	if err := rows.Err(); err != nil {
		return nil
	}

	if len(scans) == 0 {
		return nil
	}

	return fmt.Errorf(
		"query rejected: full table scan detected on a JOINed table, "+
			"problem: %s, fix: replace the JOIN with a correlated subquery",
		strings.Join(scans, "; "),
	)
}
