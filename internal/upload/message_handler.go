package upload

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
)

// messageHandler handles uploading messages to the Chat API.
type messageHandler struct {
	messages chat.Messages
	db       sqlite.Database
	logger   log.Logger
}

func newMessageHandler(messages chat.Messages, db sqlite.Database, logger log.Logger) *messageHandler {
	return &messageHandler{
		messages: messages,
		db:       db,
		logger:   logger,
	}
}

// Handle uploads a message to the backend.
func (h *messageHandler) Handle(ctx context.Context, entry *powersync.CrudEntry) error {
	switch entry.Op {
	case powersync.OpPut:
		return h.handlePut(ctx, entry)
	case powersync.OpPatch:
		// Updates to messages (e.g., streaming content) don't need upload
		// The final assistant message will be uploaded when complete
		h.logger.Debug("skipping message PATCH", "id", entry.RowID)
		return nil
	case powersync.OpDelete:
		// Messages are deleted via conversation deletion, not individually.
		// Return nil to remove from queue without API call.
		h.logger.Debug("skipping message DELETE", "id", entry.RowID)
		return nil
	default:
		h.logger.Warn("unknown message op", "op", entry.Op)
		return nil
	}
}

func (h *messageHandler) handlePut(ctx context.Context, entry *powersync.CrudEntry) error {
	role, _ := entry.Data["role"].(string)

	switch role {
	case "user":
		return h.handleUserMessage(ctx, entry)
	case "assistant":
		return h.handleAssistantMessage(ctx, entry)
	default:
		h.logger.Warn("unknown message role", "role", role)
		return nil
	}
}

// handleUserMessage uploads a user message and handles the streaming response.
func (h *messageHandler) handleUserMessage(ctx context.Context, entry *powersync.CrudEntry) error {
	conversationID, _ := entry.Data["conversation_id"].(string)
	accountID, _ := entry.Data["account_id"].(string)
	content, _ := entry.Data["content"].(string)

	// Track assistant message state
	var assistantMsgID string
	var assistantContent string
	messageCreated := false

	err := h.messages.UploadUserMessage(ctx, entry.RowID, conversationID, content, func(event chat.StreamEvent) error {
		if event.Done {
			// Stream complete - assistant message is finalized
			h.logger.Debug("stream complete", "assistantMsgID", assistantMsgID)
			return nil
		}

		// Handle different event types
		switch event.Type {
		case "message_start":
			// Create the assistant message record
			assistantMsgID = uuid.New().String()
			if err := h.createAssistantMessage(ctx, assistantMsgID, conversationID, accountID); err != nil {
				return fmt.Errorf("create assistant message: %w", err)
			}
			messageCreated = true
			h.logger.Debug("created assistant message", "id", assistantMsgID)

		case "content_block_delta", "text_delta":
			if event.Text != nil {
				assistantContent += event.Text.Content
				// Update message content for real-time display
				if messageCreated {
					if err := h.updateAssistantContent(ctx, assistantMsgID, assistantContent); err != nil {
						h.logger.Warn("failed to update assistant message", "error", err)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("upload user message: %w", err)
	}

	h.logger.Debug("uploaded user message", "id", entry.RowID)
	return nil
}

// handleAssistantMessage uploads an assistant message for durability.
func (h *messageHandler) handleAssistantMessage(ctx context.Context, entry *powersync.CrudEntry) error {
	conversationID, _ := entry.Data["conversation_id"].(string)
	content, _ := entry.Data["content"].(string)
	model, _ := entry.Data["model"].(string)
	stopReason, _ := entry.Data["stop_reason"].(string)

	err := h.messages.UploadAssistantMessage(ctx, entry.RowID, conversationID, content, model, stopReason)
	if err != nil {
		return fmt.Errorf("upload assistant message: %w", err)
	}

	h.logger.Debug("uploaded assistant message", "id", entry.RowID)
	return nil
}

// createAssistantMessage creates a new assistant message in SQLite.
func (h *messageHandler) createAssistantMessage(ctx context.Context, msgID, conversationID, accountID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	role := "assistant"
	content := "" // Will be updated as stream arrives

	return h.db.Queries().InsertMessage(ctx, sqlite.InsertMessageParams{
		ID:             &msgID,
		AccountID:      &accountID,
		ConversationID: &conversationID,
		Content:        &content,
		CreatedAt:      &now,
		Role:           &role,
	})
}

// updateAssistantContent updates an assistant message with new content.
func (h *messageHandler) updateAssistantContent(ctx context.Context, msgID, content string) error {
	return h.db.Queries().UpdateMessageContent(ctx, sqlite.UpdateMessageContentParams{
		ID:      &msgID,
		Content: &content,
	})
}
