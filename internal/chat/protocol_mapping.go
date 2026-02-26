package chat

import (
	"fmt"

	"github.com/usetero/cli/internal/domain"
)

func toWireRequest(req Request) (requestWire, error) {
	messages := make([]messageWire, 0, len(req.Messages))
	for i, msg := range req.Messages {
		wm, err := toWireMessage(msg)
		if err != nil {
			return requestWire{}, fmt.Errorf("messages[%d]: %w", i, err)
		}
		messages = append(messages, wm)
	}

	return requestWire{
		Messages: messages,
		Context:  req.Context,
		Tools:    req.Tools,
	}, nil
}

func toWireMessage(msg domain.Message) (messageWire, error) {
	blocks := make([]blockWire, 0, len(msg.Content))
	for i, b := range msg.Content {
		wb, err := toWireBlock(b)
		if err != nil {
			return messageWire{}, fmt.Errorf("content[%d]: %w", i, err)
		}
		blocks = append(blocks, wb)
	}

	return messageWire{
		Role:       msg.Role,
		Content:    blocks,
		Model:      msg.Model,
		StopReason: msg.StopReason,
	}, nil
}

func toWireBlock(b domain.Block) (blockWire, error) {
	wb := blockWire{Type: b.Type}

	switch b.Type {
	case domain.BlockTypeText:
		if b.Text == nil {
			return blockWire{}, fmt.Errorf("text block missing payload")
		}
		wb.Text = &textWire{Content: b.Text.Content}
	case domain.BlockTypeThinking:
		if b.Thinking == nil {
			return blockWire{}, fmt.Errorf("thinking block missing payload")
		}
		wb.Thinking = &thinkingWire{Content: b.Thinking.Content}
	case domain.BlockTypeToolUse:
		if b.ToolUse == nil {
			return blockWire{}, fmt.Errorf("tool_use block missing payload")
		}
		wb.ToolUse = &toolUseWire{ID: b.ToolUse.ID, Name: b.ToolUse.Name, Input: b.ToolUse.Input}
	case domain.BlockTypeToolResult:
		if b.ToolResult == nil {
			return blockWire{}, fmt.Errorf("tool_result block missing payload")
		}
		wb.ToolResult = &toolResultWire{
			ToolUseID: b.ToolResult.ToolUseID,
			IsError:   b.ToolResult.IsError,
			Error:     b.ToolResult.Error,
			Content:   b.ToolResult.Content,
		}
	default:
		return blockWire{}, fmt.Errorf("unsupported block type %q", b.Type)
	}

	return wb, nil
}
