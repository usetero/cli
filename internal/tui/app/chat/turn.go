package chat

import (
	"context"
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/app/tools"
)

// TurnEvent is sent from a Turn to the model.
type TurnEvent struct {
	// Event is a streaming event (text delta, tool use, etc).
	// Nil when Done is true.
	Event *chat.Event

	// AssistantMessage is set when an assistant response completes.
	// This may happen multiple times if there are tool calls.
	AssistantMessage *domain.Message

	// ToolResult is set when a tool has been executed.
	ToolResult *domain.Block

	// Done is true when the turn is complete (stop_reason: end_turn).
	Done bool

	// Error is set if something went wrong.
	Error error
}

// Turn handles the conversation loop: send → stream → execute tools → repeat.
type Turn interface {
	Run(ctx context.Context, conversationID string, messages []domain.Message, t tools.Tools, eventCh chan<- TurnEvent)
}

// turn is the default implementation of Turn.
type turn struct {
	client chat.Client
	logger log.Logger
}

// NewTurn creates a new Turn.
func NewTurn(client chat.Client, logger log.Logger) Turn {
	return &turn{
		client: client,
		logger: logger,
	}
}

// Run executes the turn, sending events to the channel.
// It handles the tool execution loop: send → stream → execute tools → repeat.
// Returns when stop_reason is "end_turn" or an error occurs.
func (t *turn) Run(ctx context.Context, conversationID string, messages []domain.Message, tls tools.Tools, eventCh chan<- TurnEvent) {
	defer close(eventCh)

	for {
		// Build request
		req := chat.Request{
			ConversationID: conversationID,
			Messages:       messages,
			Tools:          tls.Definitions(),
		}

		// Stream the response
		acc := chat.NewAccumulator()
		err := t.client.Send(ctx, req, func(event chat.Event) error {
			acc.Handle(event)
			eventCh <- TurnEvent{Event: &event}
			return nil
		})

		if err != nil {
			eventCh <- TurnEvent{Error: err}
			return
		}

		// Build assistant message from accumulated blocks
		assistantMsg := domain.Message{
			Role:       domain.RoleAssistant,
			Content:    acc.Blocks(),
			Model:      acc.Model(),
			StopReason: acc.StopReason(),
		}
		messages = append(messages, assistantMsg)
		eventCh <- TurnEvent{AssistantMessage: &assistantMsg}

		// Check if we're done
		if acc.StopReason() != "tool_use" {
			eventCh <- TurnEvent{Done: true}
			return
		}

		// Execute tools and collect results
		toolResults := t.executeTools(tls, acc.Blocks())
		if len(toolResults) == 0 {
			t.logger.Error("stop_reason was tool_use but no tool_use blocks found")
			eventCh <- TurnEvent{Done: true}
			return
		}

		// Send tool results as events (for UI updates)
		for i := range toolResults {
			eventCh <- TurnEvent{ToolResult: &toolResults[i]}
		}

		// Add tool results as a user message and continue the loop
		messages = append(messages, domain.Message{
			Role:    domain.RoleUser,
			Content: toolResults,
		})
	}
}

// executeTools finds tool_use blocks and executes them.
func (t *turn) executeTools(tls tools.Tools, blocks []domain.Block) []domain.Block {
	var results []domain.Block

	for _, block := range blocks {
		if block.Type != domain.BlockTypeToolUse || block.ToolUse == nil {
			continue
		}

		result := t.executeTool(tls, block.ToolUse)
		results = append(results, result)
	}

	return results
}

// executeTool executes a single tool and returns the result block.
func (t *turn) executeTool(tls tools.Tools, toolUse *domain.ToolUse) domain.Block {
	tl := tls.Get(toolUse.Name)

	if tl == nil {
		t.logger.Error("unknown tool", "name", toolUse.Name)
		return domain.Block{
			Type: domain.BlockTypeToolResult,
			ToolResult: &domain.ToolResult{
				ToolUseID: toolUse.ID,
				IsError:   true,
				Error:     "unknown tool: " + toolUse.Name,
			},
		}
	}

	// Execute the tool with the raw JSON input
	result, err := tl.Execute(toolUse.Input)
	if err != nil {
		t.logger.Error("tool execution failed", "name", toolUse.Name, "error", err)
		return domain.Block{
			Type: domain.BlockTypeToolResult,
			ToolResult: &domain.ToolResult{
				ToolUseID: toolUse.ID,
				IsError:   true,
				Error:     err.Error(),
			},
		}
	}

	// Marshal result to JSON for the content field
	content, err := json.Marshal(result)
	if err != nil {
		t.logger.Error("failed to marshal tool result", "name", toolUse.Name, "error", err)
		return domain.Block{
			Type: domain.BlockTypeToolResult,
			ToolResult: &domain.ToolResult{
				ToolUseID: toolUse.ID,
				IsError:   true,
				Error:     "failed to marshal result: " + err.Error(),
			},
		}
	}

	t.logger.Debug("tool executed", "name", toolUse.Name, "result_size", len(content))

	return domain.Block{
		Type: domain.BlockTypeToolResult,
		ToolResult: &domain.ToolResult{
			ToolUseID: toolUse.ID,
			Content:   content,
		},
	}
}
