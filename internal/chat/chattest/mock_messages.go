package chattest

import (
	"context"

	"github.com/usetero/cli/internal/chat"
)

// MockMessages implements chat.Messages for testing.
type MockMessages struct {
	UploadUserMessageFunc      func(ctx context.Context, messageID, conversationID, content string, handler chat.StreamHandler) error
	UploadAssistantMessageFunc func(ctx context.Context, messageID, conversationID, content, model, stopReason string) error
}

func (m *MockMessages) UploadUserMessage(ctx context.Context, messageID, conversationID, content string, handler chat.StreamHandler) error {
	if m.UploadUserMessageFunc != nil {
		return m.UploadUserMessageFunc(ctx, messageID, conversationID, content, handler)
	}
	return nil
}

func (m *MockMessages) UploadAssistantMessage(ctx context.Context, messageID, conversationID, content, model, stopReason string) error {
	if m.UploadAssistantMessageFunc != nil {
		return m.UploadAssistantMessageFunc(ctx, messageID, conversationID, content, model, stopReason)
	}
	return nil
}
