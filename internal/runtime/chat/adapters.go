package chat

import (
	"encoding/json"

	infrachat "github.com/usetero/cli/internal/infrastructure/chat"
)

func userTextWireMessage(text string) infrachat.Message {
	return infrachat.Message{
		Role: infrachat.RoleUser,
		Content: []infrachat.Block{{
			Type: infrachat.BlockTypeText,
			Text: &infrachat.Text{Content: text},
		}},
	}
}

func assistantTextWireMessage(text string) infrachat.Message {
	return infrachat.Message{
		Role: infrachat.RoleAssistant,
		Content: []infrachat.Block{{
			Type: infrachat.BlockTypeText,
			Text: &infrachat.Text{Content: text},
		}},
	}
}

func assistantToolUseWireMessage(toolUses ...infrachat.ToolUse) infrachat.Message {
	blocks := make([]infrachat.Block, 0, len(toolUses))
	for i := range toolUses {
		tool := toolUses[i]
		toolCopy := tool
		blocks = append(blocks, infrachat.Block{
			Type:    infrachat.BlockTypeToolUse,
			ToolUse: &toolCopy,
		})
	}
	stopReason := infrachat.StopReasonToolUse
	return infrachat.Message{
		Role:       infrachat.RoleAssistant,
		Content:    blocks,
		StopReason: &stopReason,
	}
}

func toolResultWireMessage(toolResults ...infrachat.ToolResult) infrachat.Message {
	blocks := make([]infrachat.Block, 0, len(toolResults))
	for i := range toolResults {
		result := toolResults[i]
		resultCopy := result
		blocks = append(blocks, infrachat.Block{
			Type:       infrachat.BlockTypeToolResult,
			ToolResult: &resultCopy,
		})
	}
	return infrachat.Message{
		Role:    infrachat.RoleUser,
		Content: blocks,
	}
}

func normalizeToolOutput(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	var object map[string]any
	if json.Unmarshal(raw, &object) == nil {
		return raw
	}
	return json.RawMessage(`{"result":"non_object_output"}`)
}
