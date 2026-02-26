package chat

import (
	"encoding/json"

	"github.com/usetero/cli/internal/domain"
)

// requestWire is the outbound chat request payload.
// It intentionally excludes local-only fields (message IDs, block indexes, etc).
type requestWire struct {
	Messages []messageWire          `json:"messages"`
	Context  []domain.ContextEntity `json:"context,omitempty"`
	Tools    []Tool                 `json:"tools"`
}

type messageWire struct {
	Role       domain.Role `json:"role"`
	Content    []blockWire `json:"content"`
	Model      string      `json:"model,omitempty"`
	StopReason string      `json:"stop_reason,omitempty"`
}

type blockWire struct {
	Type       domain.BlockType `json:"type"`
	Text       *textWire        `json:"text,omitempty"`
	Thinking   *thinkingWire    `json:"thinking,omitempty"`
	ToolUse    *toolUseWire     `json:"tool_use,omitempty"`
	ToolResult *toolResultWire  `json:"tool_result,omitempty"`
}

type textWire struct {
	Content string `json:"content"`
}

type thinkingWire struct {
	Content string `json:"content"`
}

type toolUseWire struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type toolResultWire struct {
	ToolUseID string         `json:"tool_use_id"`
	IsError   bool           `json:"is_error,omitempty"`
	Error     string         `json:"error,omitempty"`
	Content   map[string]any `json:"content,omitempty"`
}
