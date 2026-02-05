package domain

import "encoding/json"

// ToolUse represents the AI calling a tool.
type ToolUse struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Input         json.RawMessage `json:"input"`
	InputComplete bool            `json:"-"` // True when input is fully received (after content_block_stop)
}

// ToolResult is the outcome of a tool call (wire/storage format).
type ToolResult struct {
	ToolUseID string         `json:"tool_use_id"`
	IsError   bool           `json:"is_error,omitempty"`
	Error     string         `json:"error,omitempty"`
	Content   map[string]any `json:"content,omitempty"`
}

// ToolInputDelta is a streaming fragment of tool input JSON.
type ToolInputDelta struct {
	ToolUseID string `json:"tool_use_id"`
	Delta     string `json:"delta"`
}
