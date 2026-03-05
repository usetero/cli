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
)

// Block is one element in a message's content array.
// Exactly one of the typed fields is populated, determined by Type.
type Block struct {
	Index      int         `json:"index"`
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

// EncodeToolResults encodes tool results as blocks for storage.
func EncodeToolResults(results []ToolResult) (string, error) {
	blocks := make([]Block, len(results))
	for i := range results {
		blocks[i] = Block{
			Type:       BlockTypeToolResult,
			ToolResult: &results[i],
		}
	}
	return EncodeBlocks(blocks)
}
