package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/api/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Messages persists messages to the control plane for durability.
// This is separate from the Chat API - it only handles persistence,
// not inference.
type Messages interface {
	CreateMessage(ctx context.Context, msg *domain.Message) error
}

// MessageService handles message persistence via GraphQL.
type MessageService struct {
	client Client
	scope  log.Scope
}

var _ Messages = (*MessageService)(nil)

// NewMessageService creates a new message service.
func NewMessageService(client Client, scope log.Scope) *MessageService {
	return &MessageService{
		client: client,
		scope:  scope.Child("messages"),
	}
}

// CreateMessage persists a message to the control plane.
func (s *MessageService) CreateMessage(ctx context.Context, msg *domain.Message) error {
	s.scope.Debug("persisting message",
		"id", msg.ID.String(),
		"conversationID", msg.ConversationID.String(),
		"role", msg.Role,
	)

	content, err := toContentBlockInputs(msg.Content)
	if err != nil {
		return fmt.Errorf("convert content blocks: %w", err)
	}

	input := gen.CreateMessageInput{
		Id:             ptr(msg.ID.String()),
		ConversationID: msg.ConversationID.String(),
		Role:           toMessageRole(msg.Role),
		Content:        content,
		Model:          ptr(msg.Model),
		StopReason:     toStopReason(msg.StopReason),
	}

	_, err = s.client.CreateMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	return nil
}

func toMessageRole(role domain.Role) gen.MessageRole {
	switch role {
	case domain.RoleUser:
		return gen.MessageRoleUser
	case domain.RoleAssistant:
		return gen.MessageRoleAssistant
	default:
		return gen.MessageRoleUser
	}
}

func toStopReason(reason string) *gen.MessageStopReason {
	switch reason {
	case "end_turn":
		return ptr(gen.MessageStopReasonEndTurn)
	case "tool_use":
		return ptr(gen.MessageStopReasonToolUse)
	default:
		return nil // Don't send stopReason for user messages
	}
}

func toContentBlockInputs(blocks []domain.Block) ([]gen.ContentBlockInput, error) {
	inputs := make([]gen.ContentBlockInput, 0, len(blocks))

	for _, block := range blocks {
		input, err := toContentBlockInput(block)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}

func toContentBlockInput(block domain.Block) (gen.ContentBlockInput, error) {
	switch block.Type {
	case domain.BlockTypeText:
		return gen.ContentBlockInput{
			Type: gen.ContentBlockTypeText,
			Text: &gen.TextBlockInput{
				Content: block.Text.Content,
			},
		}, nil

	case domain.BlockTypeThinking:
		return gen.ContentBlockInput{
			Type: gen.ContentBlockTypeThinking,
			Thinking: &gen.ThinkingBlockInput{
				Content: block.Thinking.Content,
			},
		}, nil

	case domain.BlockTypeToolUse:
		input, err := rawJSONToMap(block.ToolUse.Input)
		if err != nil {
			return gen.ContentBlockInput{}, fmt.Errorf("tool_use block: %w", err)
		}
		return gen.ContentBlockInput{
			Type: gen.ContentBlockTypeToolUse,
			ToolUse: &gen.ToolUseInput{
				Id:    block.ToolUse.ID,
				Name:  block.ToolUse.Name,
				Input: input,
			},
		}, nil

	case domain.BlockTypeToolResult:
		var content *map[string]any
		if block.ToolResult.Content != nil {
			content = &block.ToolResult.Content
		}
		toolResult := &gen.ToolResultInput{
			ToolUseId: block.ToolResult.ToolUseID,
			IsError:   block.ToolResult.IsError,
			Content:   content,
		}
		if block.ToolResult.Error != "" {
			toolResult.Error = ptr(block.ToolResult.Error)
		}
		return gen.ContentBlockInput{
			Type:       gen.ContentBlockTypeToolResult,
			ToolResult: toolResult,
		}, nil

	default:
		return gen.ContentBlockInput{}, fmt.Errorf("unknown block type: %s", block.Type)
	}
}

func rawJSONToMap(data json.RawMessage) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}
