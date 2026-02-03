package domain

import (
	"encoding/json"
	"fmt"
)

// BlockType identifies the kind of content block.
type BlockType string

const (
	// Persistable block types - stored in the database.
	BlockTypeText       BlockType = "text"
	BlockTypeThinking   BlockType = "thinking"
	BlockTypeToolUse    BlockType = "tool_use"
	BlockTypeToolResult BlockType = "tool_result"

	// Streaming-only block types - sent via SSE, never persisted.
	BlockTypeTextDelta      BlockType = "text_delta"
	BlockTypeThinkingDelta  BlockType = "thinking_delta"
	BlockTypeToolInputDelta BlockType = "tool_input_delta"

	// Stream control types - signal stream lifecycle events.
	BlockTypeMessageStart BlockType = "message_start"
	BlockTypeMessageStop  BlockType = "message_stop"
)

// Block is one element in a message's content array.
// Exactly one of the typed fields is populated, determined by Type.
type Block struct {
	Type           BlockType     `json:"type"`
	Text           *TextBlock    `json:"text,omitempty"`
	Thinking       *Thinking     `json:"thinking,omitempty"`
	ToolUse        *ToolUse      `json:"tool_use,omitempty"`
	ToolResult     *ToolResult   `json:"tool_result,omitempty"`
	ToolInputDelta string        `json:"tool_input_delta,omitempty"`
	MessageStart   *MessageStart `json:"message_start,omitempty"`
	MessageStop    *MessageStop  `json:"message_stop,omitempty"`
}

// TextBlock is prose content.
type TextBlock struct {
	Content string `json:"content"`
}

// Thinking is the AI's internal reasoning.
type Thinking struct {
	Content string `json:"content"`
}

// MessageStart contains metadata sent at the start of a message stream.
type MessageStart struct {
	Model string `json:"model"`
}

// MessageStop contains metadata sent at the end of a message stream.
type MessageStop struct {
	StopReason string `json:"stop_reason"`
}

// IsDelta returns true if the block is a streaming-only delta type.
func (b Block) IsDelta() bool {
	return b.Type == BlockTypeTextDelta || b.Type == BlockTypeThinkingDelta || b.Type == BlockTypeToolInputDelta
}

// ParseBlocks parses a JSON string into content blocks.
func ParseBlocks(data string) ([]Block, error) {
	if data == "" {
		return nil, nil
	}
	var blocks []Block
	if err := json.Unmarshal([]byte(data), &blocks); err != nil {
		return nil, fmt.Errorf("parse blocks: %w", err)
	}
	return blocks, nil
}

// EncodeBlocks serializes blocks to JSON string for storage.
func EncodeBlocks(blocks []Block) (string, error) {
	if len(blocks) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(blocks)
	if err != nil {
		return "", fmt.Errorf("encode blocks: %w", err)
	}
	return string(data), nil
}

// NewTextBlock creates a text block.
func NewTextBlock(content string) Block {
	return Block{
		Type: BlockTypeText,
		Text: &TextBlock{Content: content},
	}
}

// EncodeText is a convenience function to encode a single text block.
func EncodeText(content string) (string, error) {
	return EncodeBlocks([]Block{NewTextBlock(content)})
}
