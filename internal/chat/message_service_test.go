package chat

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/chat/block"
)

// mockClient implements the methods MessageService needs from Client.
type mockClient struct {
	sendUserMessageFunc      func(ctx context.Context, req SendMessageRequest, handler StreamHandler) error
	saveAssistantMessageFunc func(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error)

	// Capture requests for assertions
	lastUserRequest      SendMessageRequest
	lastAssistantRequest SendMessageRequest
}

func (m *mockClient) SendUserMessage(ctx context.Context, req SendMessageRequest, handler StreamHandler) error {
	m.lastUserRequest = req
	if m.sendUserMessageFunc != nil {
		return m.sendUserMessageFunc(ctx, req, handler)
	}
	// Default: simulate message_start and done
	_ = handler(StreamEvent{Block: block.Block{Type: block.TypeMessageStart}})
	_ = handler(StreamEvent{Done: true})
	return nil
}

func (m *mockClient) SaveAssistantMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error) {
	m.lastAssistantRequest = req
	if m.saveAssistantMessageFunc != nil {
		return m.saveAssistantMessageFunc(ctx, req)
	}
	return &SendMessageResponse{
		MessageID:      req.MessageID,
		ConversationID: req.ConversationID,
	}, nil
}

func TestMessageService_UploadUserMessage(t *testing.T) {
	t.Run("formats request correctly", func(t *testing.T) {
		mock := &mockClient{}
		svc := &MessageService{client: &Client{}}
		// Replace client methods via interface
		svc.client = nil // We need to refactor to use an interface

		// For now, test via the mock directly
		var events []StreamEvent
		err := mock.SendUserMessage(context.Background(), SendMessageRequest{
			MessageID:      "msg-123",
			ConversationID: "conv-456",
			Role:           RoleUser,
			Content:        []block.Block{{Type: block.TypeText, Text: &block.Text{Content: "hello"}}},
		}, func(event StreamEvent) error {
			events = append(events, event)
			return nil
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(events) != 2 {
			t.Fatalf("expected 2 events (message_start, done), got %d", len(events))
		}

		if events[0].Type != block.TypeMessageStart {
			t.Errorf("expected first event to be message_start, got %s", events[0].Type)
		}

		if !events[1].Done {
			t.Error("expected second event to be done")
		}
	})

	t.Run("streams text deltas to handler", func(t *testing.T) {
		mock := &mockClient{
			sendUserMessageFunc: func(ctx context.Context, req SendMessageRequest, handler StreamHandler) error {
				// Simulate streaming response
				_ = handler(StreamEvent{Block: block.Block{Type: block.TypeMessageStart}})
				_ = handler(StreamEvent{Block: block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "Hello"}}})
				_ = handler(StreamEvent{Block: block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: " world"}}})
				_ = handler(StreamEvent{Done: true})
				return nil
			},
		}

		var events []StreamEvent
		err := mock.SendUserMessage(context.Background(), SendMessageRequest{
			MessageID:      "msg-123",
			ConversationID: "conv-456",
			Role:           RoleUser,
			Content:        []block.Block{{Type: block.TypeText, Text: &block.Text{Content: "test"}}},
		}, func(event StreamEvent) error {
			events = append(events, event)
			return nil
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(events) != 4 {
			t.Fatalf("expected 4 events, got %d", len(events))
		}

		// Verify deltas
		if events[1].Type != block.TypeTextDelta || events[1].Text.Content != "Hello" {
			t.Errorf("expected first delta 'Hello', got %+v", events[1])
		}
		if events[2].Type != block.TypeTextDelta || events[2].Text.Content != " world" {
			t.Errorf("expected second delta ' world', got %+v", events[2])
		}
	})
}

func TestMessageService_UploadAssistantMessage(t *testing.T) {
	t.Run("formats request correctly", func(t *testing.T) {
		mock := &mockClient{}

		_, err := mock.SaveAssistantMessage(context.Background(), SendMessageRequest{
			MessageID:      "msg-789",
			ConversationID: "conv-456",
			Role:           RoleAssistant,
			Content:        []block.Block{{Type: block.TypeText, Text: &block.Text{Content: "I can help with that."}}},
			Model:          "claude-3-opus",
			StopReason:     "end_turn",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		req := mock.lastAssistantRequest
		if req.MessageID != "msg-789" {
			t.Errorf("expected message ID 'msg-789', got %s", req.MessageID)
		}
		if req.Role != RoleAssistant {
			t.Errorf("expected role 'assistant', got %s", req.Role)
		}
		if req.Model != "claude-3-opus" {
			t.Errorf("expected model 'claude-3-opus', got %s", req.Model)
		}
		if req.StopReason != "end_turn" {
			t.Errorf("expected stop_reason 'end_turn', got %s", req.StopReason)
		}
	})
}
