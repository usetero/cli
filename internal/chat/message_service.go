// Package chat provides the Chat API client and message upload services.
// This package handles communication with the Tero Chat API.
// Local SQLite writes are handled by the caller (upload loop).
package chat

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/chat/block"
)

// Messages provides access to message operations.
type Messages interface {
	UploadUserMessage(ctx context.Context, messageID, conversationID, content string, handler StreamHandler) error
	UploadAssistantMessage(ctx context.Context, messageID, conversationID, content, model, stopReason string) error
}

// Ensure MessageService implements Messages.
var _ Messages = (*MessageService)(nil)

// MessageService handles uploading messages to the Chat API.
type MessageService struct {
	client *Client
}

// NewMessageService creates a new message service.
func NewMessageService(client *Client) *MessageService {
	return &MessageService{
		client: client,
	}
}

// UploadUserMessage uploads a user message to the Chat API and streams back
// the assistant response. The handler is called for each stream event.
// The caller is responsible for writing the assistant message to SQLite.
func (s *MessageService) UploadUserMessage(ctx context.Context, messageID, conversationID, content string, handler StreamHandler) error {
	req := SendMessageRequest{
		MessageID:      messageID,
		ConversationID: conversationID,
		Role:           RoleUser,
		Content: []block.Block{
			{Type: block.TypeText, Text: &block.Text{Content: content}},
		},
	}

	err := s.client.SendUserMessage(ctx, req, handler)
	if err != nil {
		return fmt.Errorf("upload user message: %w", err)
	}

	return nil
}

// UploadAssistantMessage uploads an assistant message to the Chat API for durability.
func (s *MessageService) UploadAssistantMessage(ctx context.Context, messageID, conversationID, content, model, stopReason string) error {
	req := SendMessageRequest{
		MessageID:      messageID,
		ConversationID: conversationID,
		Role:           RoleAssistant,
		Content: []block.Block{
			{Type: block.TypeText, Text: &block.Text{Content: content}},
		},
		Model:      model,
		StopReason: stopReason,
	}

	_, err := s.client.SaveAssistantMessage(ctx, req)
	if err != nil {
		return fmt.Errorf("upload assistant message: %w", err)
	}

	return nil
}
