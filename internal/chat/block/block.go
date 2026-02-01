// Package block defines content block types for chat messages.
// Content blocks are the building blocks of message content - text, thinking,
// tool calls, and tool results. This package handles both persisted block types
// and streaming delta types.
package block

import (
	"encoding/json"
	"fmt"
)

// Type identifies the kind of content block.
type Type string

const (
	// Persistable block types - stored in the database.
	TypeText       Type = "text"
	TypeThinking   Type = "thinking"
	TypeToolUse    Type = "tool_use"
	TypeToolResult Type = "tool_result"

	// Streaming-only block types - sent via SSE, never persisted.
	TypeTextDelta      Type = "text_delta"
	TypeThinkingDelta  Type = "thinking_delta"
	TypeToolInputDelta Type = "tool_input_delta"

	// Stream control types - signal stream lifecycle events.
	TypeMessageStart Type = "message_start"
)

// Block is one element in a message's content array.
// Exactly one of the typed fields is populated, determined by Type.
type Block struct {
	Type           Type        `json:"type"`
	Text           *Text       `json:"text,omitempty"`
	Thinking       *Thinking   `json:"thinking,omitempty"`
	ToolUse        *ToolUse    `json:"tool_use,omitempty"`
	ToolResult     *ToolResult `json:"tool_result,omitempty"`
	ToolInputDelta string      `json:"tool_input_delta,omitempty"` // Streaming only - partial JSON
}

// Validate checks that exactly one typed field is populated and matches the type.
func (b Block) Validate() error {
	switch b.Type {
	case TypeText, TypeTextDelta:
		if b.Text == nil {
			return fmt.Errorf("text block requires text field")
		}
	case TypeThinking, TypeThinkingDelta:
		if b.Thinking == nil {
			return fmt.Errorf("thinking block requires thinking field")
		}
	case TypeToolUse:
		if b.ToolUse == nil {
			return fmt.Errorf("tool_use block requires tool_use field")
		}
	case TypeToolResult:
		if b.ToolResult == nil {
			return fmt.Errorf("tool_result block requires tool_result field")
		}
	case "":
		return fmt.Errorf("type is required")
	default:
		return fmt.Errorf("unknown block type %q", b.Type)
	}
	return nil
}

// IsDelta returns true if the block is a streaming-only delta type.
func (b Block) IsDelta() bool {
	return b.Type == TypeTextDelta || b.Type == TypeThinkingDelta || b.Type == TypeToolInputDelta
}

// Parse parses a JSON string into content blocks.
// Returns nil for empty input.
func Parse(data string) ([]Block, error) {
	if data == "" {
		return nil, nil
	}
	var blocks []Block
	if err := json.Unmarshal([]byte(data), &blocks); err != nil {
		return nil, err
	}
	return blocks, nil
}

// Encode serializes blocks to JSON string for storage.
func Encode(blocks []Block) (string, error) {
	if len(blocks) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(blocks)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
