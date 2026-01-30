package chat

import (
	"encoding/json"
)

// BlockType identifies the kind of content block.
type BlockType string

const (
	BlockTypeText       BlockType = "text"
	BlockTypeThinking   BlockType = "thinking"
	BlockTypeToolUse    BlockType = "tool_use"
	BlockTypeToolResult BlockType = "tool_result"
)

// ContentBlock is one element in a message's content array.
// Mirrors the control plane's schema.ContentBlock for JSON compatibility.
type ContentBlock struct {
	Type       BlockType   `json:"type"`
	Text       *TextBlock  `json:"text,omitempty"`
	Thinking   *Thinking   `json:"thinking,omitempty"`
	ToolUse    *ToolUse    `json:"tool_use,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// TextBlock is prose content.
type TextBlock struct {
	Content string `json:"content"`
}

// Thinking is the AI's internal reasoning.
type Thinking struct {
	Content string `json:"content"`
}

// ToolUse represents a tool call from the AI.
type ToolUse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Input is stored as raw JSON since tool inputs vary by tool type
	Input json.RawMessage `json:"input,omitempty"`
}

// ToolResult is the result of executing a tool.
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	IsError   bool   `json:"is_error,omitempty"`
	Error     string `json:"error,omitempty"`
	// Content is the result payload, varies by tool type
	Content json.RawMessage `json:"content,omitempty"`
}

// ParseContentBlocks parses a JSON string into content blocks.
func ParseContentBlocks(data string) ([]ContentBlock, error) {
	if data == "" {
		return nil, nil
	}
	var blocks []ContentBlock
	if err := json.Unmarshal([]byte(data), &blocks); err != nil {
		return nil, err
	}
	return blocks, nil
}
