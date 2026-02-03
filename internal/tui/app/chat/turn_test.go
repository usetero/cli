package chat_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/chattest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	appchat "github.com/usetero/cli/internal/tui/app/chat"
	"github.com/usetero/cli/internal/tui/app/tools"
	"github.com/usetero/cli/internal/tui/app/tools/toolstest"
)

func TestTurn_Run(t *testing.T) {
	t.Parallel()

	t.Run("sends events and Done on end_turn", func(t *testing.T) {
		t.Parallel()

		client := &chattest.MockClient{
			SendFunc: func(ctx context.Context, req chat.Request, handler chat.Handler) error {
				// Simulate streaming a text response
				handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeMessageStart, MessageStart: &domain.MessageStart{Model: "test-model"}}})
				handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeTextDelta, Text: &domain.TextBlock{Content: "Hello"}}})
				handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeMessageStop, MessageStop: &domain.MessageStop{StopReason: "end_turn"}}})
				handler(chat.Event{Done: true})
				return nil
			},
		}

		turn := appchat.NewTurn(client, logtest.New(t))
		eventCh := make(chan appchat.TurnEvent, 100)

		go turn.Run(context.Background(), "conv-1", nil, nil, eventCh)

		var events []appchat.TurnEvent
		for ev := range eventCh {
			events = append(events, ev)
		}

		// Should have streaming events + assistant message + done
		if len(events) < 3 {
			t.Fatalf("expected at least 3 events, got %d", len(events))
		}

		// Last event should be Done
		last := events[len(events)-1]
		if !last.Done {
			t.Error("expected last event to have Done=true")
		}

		// Second to last should be AssistantMessage
		secondLast := events[len(events)-2]
		if secondLast.AssistantMessage == nil {
			t.Error("expected AssistantMessage before Done")
		}
	})

	t.Run("executes tools and loops on tool_use", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		client := &chattest.MockClient{
			SendFunc: func(ctx context.Context, req chat.Request, handler chat.Handler) error {
				callCount++
				if callCount == 1 {
					// First call: respond with tool_use
					handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeMessageStart, MessageStart: &domain.MessageStart{Model: "test-model"}}})
					handler(chat.Event{Block: domain.Block{
						Type: domain.BlockTypeToolUse,
						ToolUse: &domain.ToolUse{
							ID:    "tool-1",
							Name:  "test_tool",
							Input: json.RawMessage(`{"foo":"bar"}`),
						},
					}})
					handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeMessageStop, MessageStop: &domain.MessageStop{StopReason: "tool_use"}}})
					handler(chat.Event{Done: true})
				} else {
					// Second call: should have tool result in messages, respond with end_turn
					foundToolResult := false
					for _, msg := range req.Messages {
						for _, block := range msg.Content {
							if block.Type == domain.BlockTypeToolResult {
								foundToolResult = true
							}
						}
					}
					if !foundToolResult {
						t.Error("expected tool result in messages on second call")
					}

					handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeMessageStart, MessageStart: &domain.MessageStart{Model: "test-model"}}})
					handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeTextDelta, Text: &domain.TextBlock{Content: "Done"}}})
					handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeMessageStop, MessageStop: &domain.MessageStop{StopReason: "end_turn"}}})
					handler(chat.Event{Done: true})
				}
				return nil
			},
		}

		testTool := toolstest.MockTool{
			NameVal: "test_tool",
			ExecuteFn: func(input json.RawMessage) (any, error) {
				return map[string]string{"result": "success"}, nil
			},
		}

		turn := appchat.NewTurn(client, logtest.New(t))
		eventCh := make(chan appchat.TurnEvent, 100)

		go turn.Run(context.Background(), "conv-1", nil, tools.Tools{testTool}, eventCh)

		var events []appchat.TurnEvent
		for ev := range eventCh {
			events = append(events, ev)
		}

		if callCount != 2 {
			t.Errorf("expected 2 client calls (tool_use then end_turn), got %d", callCount)
		}

		// Should have tool result event
		hasToolResult := false
		for _, ev := range events {
			if ev.ToolResult != nil {
				hasToolResult = true
				if ev.ToolResult.ToolResult.IsError {
					t.Errorf("tool result should not be error")
				}
			}
		}
		if !hasToolResult {
			t.Error("expected ToolResult event")
		}
	})

	t.Run("sends error event on client error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("connection failed")
		client := &chattest.MockClient{
			SendFunc: func(ctx context.Context, req chat.Request, handler chat.Handler) error {
				return expectedErr
			},
		}

		turn := appchat.NewTurn(client, logtest.New(t))
		eventCh := make(chan appchat.TurnEvent, 100)

		go turn.Run(context.Background(), "conv-1", nil, nil, eventCh)

		var events []appchat.TurnEvent
		for ev := range eventCh {
			events = append(events, ev)
		}

		if len(events) != 1 {
			t.Fatalf("expected 1 event (error), got %d", len(events))
		}

		if events[0].Error == nil {
			t.Fatal("expected error event")
		}
		if events[0].Error.Error() != expectedErr.Error() {
			t.Errorf("error = %v, want %v", events[0].Error, expectedErr)
		}
	})

	t.Run("returns error for unknown tool", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		client := &chattest.MockClient{
			SendFunc: func(ctx context.Context, req chat.Request, handler chat.Handler) error {
				callCount++
				if callCount == 1 {
					// First call - request unknown tool
					handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeMessageStart, MessageStart: &domain.MessageStart{Model: "test-model"}}})
					handler(chat.Event{Block: domain.Block{
						Type: domain.BlockTypeToolUse,
						ToolUse: &domain.ToolUse{
							ID:   "tool-1",
							Name: "unknown_tool",
						},
					}})
					handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeMessageStop, MessageStop: &domain.MessageStop{StopReason: "tool_use"}}})
					handler(chat.Event{Done: true})
				} else {
					// Second call - end the conversation
					handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeMessageStart, MessageStart: &domain.MessageStart{Model: "test-model"}}})
					handler(chat.Event{Block: domain.Block{Type: domain.BlockTypeMessageStop, MessageStop: &domain.MessageStop{StopReason: "end_turn"}}})
					handler(chat.Event{Done: true})
				}
				return nil
			},
		}

		turn := appchat.NewTurn(client, logtest.New(t))
		eventCh := make(chan appchat.TurnEvent, 100)

		go turn.Run(context.Background(), "conv-1", nil, tools.Tools{}, eventCh)

		var toolResultEvent *appchat.TurnEvent
		for ev := range eventCh {
			if ev.ToolResult != nil {
				toolResultEvent = &ev
			}
		}

		if toolResultEvent == nil {
			t.Fatal("expected tool result event")
		}

		if !toolResultEvent.ToolResult.ToolResult.IsError {
			t.Error("expected tool result to be an error")
		}

		if toolResultEvent.ToolResult.ToolResult.Error == "" {
			t.Error("expected error message")
		}
	})
}
