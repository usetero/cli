package domain

import "encoding/json"

// ToolUse represents the AI calling a tool.
type ToolUse struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult is the outcome of a tool call.
type ToolResult struct {
	ToolUseID string          `json:"tool_use_id"`
	IsError   bool            `json:"is_error,omitempty"`
	Error     string          `json:"error,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
}

// ToolInputDelta is a streaming fragment of tool input JSON.
type ToolInputDelta struct {
	ToolUseID string `json:"tool_use_id"`
	Delta     string `json:"delta"`
}
