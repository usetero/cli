package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/api/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/domain/tool"
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
	logger log.Logger
}

var _ Messages = (*MessageService)(nil)

// NewMessageService creates a new message service.
func NewMessageService(client Client, logger log.Logger) *MessageService {
	return &MessageService{
		client: client,
		logger: logger,
	}
}

// CreateMessage persists a message to the control plane.
func (s *MessageService) CreateMessage(ctx context.Context, msg *domain.Message) error {
	s.logger.Debug("persisting message",
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
		input, err := toToolInput(block.ToolUse)
		if err != nil {
			return gen.ContentBlockInput{}, fmt.Errorf("tool_use block: %w", err)
		}
		return gen.ContentBlockInput{
			Type: gen.ContentBlockTypeToolUse,
			ToolUse: &gen.ToolUseInput{
				Id:    block.ToolUse.ID,
				Name:  string(block.ToolUse.Name),
				Input: input,
			},
		}, nil

	case domain.BlockTypeToolResult:
		content, err := toToolResultContent(block.ToolResult)
		if err != nil {
			return gen.ContentBlockInput{}, fmt.Errorf("tool_result block: %w", err)
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

	case domain.BlockTypeTextDelta,
		domain.BlockTypeThinkingDelta,
		domain.BlockTypeToolInputDelta,
		domain.BlockTypeMessageStart,
		domain.BlockTypeMessageStop:
		// These are streaming-only types that should never be persisted
		return gen.ContentBlockInput{}, fmt.Errorf("cannot persist streaming block type: %s", block.Type)
	}

	return gen.ContentBlockInput{}, fmt.Errorf("unknown block type: %s", block.Type)
}

func toToolInput(use *tool.Use) (map[string]any, error) {
	switch use.Name {
	case tool.AddContext:
		if use.AddContext == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.AddContext)
	case tool.RemoveContext:
		if use.RemoveContext == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.RemoveContext)
	case tool.Query:
		if use.Query == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.Query)
	case tool.ShowMetric:
		if use.ShowMetric == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.ShowMetric)
	case tool.ShowSeries:
		if use.ShowSeries == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.ShowSeries)
	case tool.ShowTimeSeries:
		if use.ShowTimeSeries == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.ShowTimeSeries)
	case tool.ShowTable:
		if use.ShowTable == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.ShowTable)
	case tool.StartJourney:
		if use.StartJourney == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.StartJourney)
	case tool.EndJourney:
		if use.EndJourney == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.EndJourney)
	case tool.ApprovePolicy:
		if use.ApprovePolicy == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.ApprovePolicy)
	case tool.DismissPolicy:
		if use.DismissPolicy == nil {
			return nil, fmt.Errorf("tool %s: missing typed input", use.Name)
		}
		return structToMap(use.DismissPolicy)
	}

	return nil, fmt.Errorf("unknown tool: %s", use.Name)
}

func toToolResultContent(result *tool.Result) (*map[string]any, error) {
	if result.AddContext != nil {
		m, err := structToMap(result.AddContext)
		return &m, err
	}
	if result.RemoveContext != nil {
		m, err := structToMap(result.RemoveContext)
		return &m, err
	}
	if result.Query != nil {
		m, err := structToMap(result.Query)
		return &m, err
	}
	if result.ApprovePolicy != nil {
		m, err := structToMap(result.ApprovePolicy)
		return &m, err
	}
	if result.DismissPolicy != nil {
		m, err := structToMap(result.DismissPolicy)
		return &m, err
	}

	// No typed result - this is OK for client-executed tools
	return nil, nil
}

func structToMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal struct: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal to map: %w", err)
	}
	return m, nil
}
