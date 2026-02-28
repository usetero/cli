package chat

import (
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/api/chatclient/protocolv2"
	"github.com/usetero/cli/internal/domain"
)

func toWireRequest(req Request) (protocolv2.Request, error) {
	messages := make([]protocolv2.Message, 0, len(req.Messages))
	for i, msg := range req.Messages {
		wm, err := toWireMessage(msg)
		if err != nil {
			return protocolv2.Request{}, fmt.Errorf("messages[%d]: %w", i, err)
		}
		messages = append(messages, wm)
	}

	wire := protocolv2.Request{
		ChatProtocolVersion: protocolv2.Version,
		ConversationID:      req.ConversationID,
		Messages:            messages,
		ContextEntities:     toWireContextEntities(req.ContextEntities),
		Tools:               toWireTools(req.Tools),
	}
	if err := protocolv2.Validate(wire); err != nil {
		return protocolv2.Request{}, err
	}
	return wire, nil
}

func toWireMessage(msg domain.Message) (protocolv2.Message, error) {
	role, err := mapWireRole(msg.Role)
	if err != nil {
		return protocolv2.Message{}, err
	}
	stopReason, err := mapWireStopReason(msg.StopReason)
	if err != nil {
		return protocolv2.Message{}, err
	}

	blocks := make([]protocolv2.Block, 0, len(msg.Content))
	for i, b := range msg.Content {
		wb, err := toWireBlock(b)
		if err != nil {
			return protocolv2.Message{}, fmt.Errorf("content[%d]: %w", i, err)
		}
		blocks = append(blocks, wb)
	}

	return protocolv2.Message{
		Role:       role,
		Content:    blocks,
		Model:      msg.Model,
		StopReason: stopReason,
	}, nil
}

func toWireBlock(b domain.Block) (protocolv2.Block, error) {
	blockType, err := mapWireBlockType(b.Type)
	if err != nil {
		return protocolv2.Block{}, err
	}
	wb := protocolv2.Block{Type: blockType}

	switch b.Type {
	case domain.BlockTypeText:
		if b.Text == nil {
			return protocolv2.Block{}, fmt.Errorf("text block missing payload")
		}
		wb.Text = &protocolv2.Text{Content: b.Text.Content}
	case domain.BlockTypeThinking:
		if b.Thinking == nil {
			return protocolv2.Block{}, fmt.Errorf("thinking block missing payload")
		}
		wb.Thinking = &protocolv2.Thinking{Content: b.Thinking.Content}
	case domain.BlockTypeToolUse:
		if b.ToolUse == nil {
			return protocolv2.Block{}, fmt.Errorf("tool_use block missing payload")
		}
		wb.ToolUse = &protocolv2.ToolUse{ID: b.ToolUse.ID, Name: b.ToolUse.Name, Input: b.ToolUse.Input}
	case domain.BlockTypeToolResult:
		if b.ToolResult == nil {
			return protocolv2.Block{}, fmt.Errorf("tool_result block missing payload")
		}
		wb.ToolResult = &protocolv2.ToolResult{
			ToolUseID: b.ToolResult.ToolUseID,
			IsError:   boolPtr(b.ToolResult.IsError),
			Error:     b.ToolResult.Error,
			Content:   sanitizeToolResultContent(b.ToolResult.Content, b.ToolResult.IsError),
		}
	default:
		return protocolv2.Block{}, fmt.Errorf("unsupported block type %q", b.Type)
	}

	return wb, nil
}

func toWireContextEntities(entities []domain.ContextEntity) []protocolv2.ContextEntity {
	out := make([]protocolv2.ContextEntity, 0, len(entities))
	for _, entity := range entities {
		out = append(out, protocolv2.ContextEntity{
			EntityType: protocolv2.ContextEntityType(entity.EntityType),
			EntityID:   entity.EntityID,
		})
	}
	return out
}

func toWireTools(tools []Tool) []protocolv2.Tool {
	out := make([]protocolv2.Tool, 0, len(tools))
	for _, tool := range tools {
		out = append(out, protocolv2.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: schemaToMap(tool.InputSchema),
		})
	}
	return out
}

func sanitizeToolResultContent(content map[string]any, isError bool) json.RawMessage {
	if isError {
		return nil
	}

	if len(content) == 0 {
		return json.RawMessage(`{}`)
	}

	out := make(map[string]any, len(content))
	for k, v := range content {
		switch k {
		case "tool_use_id", "is_error", "error":
			continue
		default:
			out[k] = v
		}
	}
	if len(out) == 0 {
		return json.RawMessage(`{}`)
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return encoded
}

func boolPtr(v bool) *bool {
	return &v
}

func schemaToMap(schema Schema) map[string]any {
	encoded, err := json.Marshal(schema)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(encoded, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func mapWireRole(role domain.Role) (protocolv2.Role, error) {
	switch role {
	case domain.RoleUser:
		return protocolv2.RoleUser, nil
	case domain.RoleAssistant:
		return protocolv2.RoleAssistant, nil
	default:
		return "", fmt.Errorf("unsupported role %q", role)
	}
}

func mapWireStopReason(reason string) (*protocolv2.StopReason, error) {
	switch reason {
	case "":
		return nil, nil
	case string(protocolv2.StopReasonEndTurn):
		v := protocolv2.StopReasonEndTurn
		return &v, nil
	case string(protocolv2.StopReasonToolUse):
		v := protocolv2.StopReasonToolUse
		return &v, nil
	default:
		return nil, fmt.Errorf("unsupported stop_reason %q", reason)
	}
}

func mapWireBlockType(blockType domain.BlockType) (protocolv2.BlockType, error) {
	switch blockType {
	case domain.BlockTypeText:
		return protocolv2.BlockTypeText, nil
	case domain.BlockTypeThinking:
		return protocolv2.BlockTypeThinking, nil
	case domain.BlockTypeToolUse:
		return protocolv2.BlockTypeToolUse, nil
	case domain.BlockTypeToolResult:
		return protocolv2.BlockTypeToolResult, nil
	default:
		return "", fmt.Errorf("unsupported block type %q", blockType)
	}
}
