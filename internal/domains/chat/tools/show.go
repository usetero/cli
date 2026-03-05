package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

type ShowTool struct {
	db *sqlite.DB
}

type ShowInput struct {
	Entity string `json:"entity"`
	ID     string `json:"id,omitempty"`
	SQL    string `json:"sql,omitempty"`
	Title  string `json:"title,omitempty"`
}

type ShowResult struct {
	Entity string `json:"entity"`
	ID     string `json:"id"`
	Title  string `json:"title,omitempty"`
}

func NewShowTool(db *sqlite.DB) *ShowTool {
	return &ShowTool{db: db}
}

func (t *ShowTool) Definition() Definition {
	return Definition{
		Name:        ShowToolName,
		Description: "Resolve one entity by id or lookup SQL and present it.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"entity": map[string]any{"type": "string"},
				"id":     map[string]any{"type": "string"},
				"sql":    map[string]any{"type": "string"},
				"title":  map[string]any{"type": "string"},
			},
			"required": []string{"entity"},
		},
	}
}

func (t *ShowTool) Run(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	if t == nil || t.db == nil {
		return nil, fmt.Errorf("show tool is not initialized")
	}

	var in ShowInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("parse show input: %w", err)
	}
	in.Entity = strings.TrimSpace(in.Entity)
	if in.Entity == "" {
		return nil, fmt.Errorf("entity is required")
	}

	id := strings.TrimSpace(in.ID)
	if id == "" {
		sqlText := strings.TrimSpace(in.SQL)
		if sqlText == "" {
			return nil, fmt.Errorf("either id or sql is required")
		}
		var err error
		id, err = t.resolveIDFromSQL(ctx, sqlText)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(ShowResult{
		Entity: in.Entity,
		ID:     id,
		Title:  strings.TrimSpace(in.Title),
	})
}

func (t *ShowTool) resolveIDFromSQL(ctx context.Context, sqlText string) (string, error) {
	qctx, cancel := sqlite.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := t.db.Raw().QueryContext(qctx, sqlText)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}
	idIdx := -1
	for i, col := range cols {
		if strings.EqualFold(col, "id") {
			idIdx = i
			break
		}
	}
	if idIdx < 0 {
		return "", fmt.Errorf("sql must return an id column")
	}

	if !rows.Next() {
		return "", fmt.Errorf("sql returned no rows")
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
		return "", fmt.Errorf("sql must return exactly one row")
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	switch v := values[idIdx].(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", fmt.Errorf("id column must be string")
	}
}
