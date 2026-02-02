package upload

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
)

// Message handler events

// MessageProcessingEvent is emitted when we start processing a user message.
// The TUI can use this to show a loading indicator until the assistant message appears.
type MessageProcessingEvent struct {
	ConversationID string
	UserMessageID  string
}

func (MessageProcessingEvent) uploadEvent() {}

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
func (h *messageHandler) Handle(ctx context.Context, entry *powersync.CrudEntry, emit Emitter) error {
	switch entry.Op {
	case powersync.OpPut:
		return h.handlePut(ctx, entry, emit)
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

func (h *messageHandler) handlePut(ctx context.Context, entry *powersync.CrudEntry, emit Emitter) error {
	role, _ := entry.Data["role"].(string)

	switch chat.MessageRole(role) {
	case chat.RoleUser:
		return h.handleUserMessage(ctx, entry, emit)
	case chat.RoleAssistant:
		return h.handleAssistantMessage(ctx, entry)
	default:
		h.logger.Warn("unknown message role", "role", role)
		return nil
	}
}

// handleUserMessage uploads a user message and handles the streaming response.
func (h *messageHandler) handleUserMessage(ctx context.Context, entry *powersync.CrudEntry, emit Emitter) error {
	conversationID, _ := entry.Data["conversation_id"].(string)
	accountID, _ := entry.Data["account_id"].(string)
	content, _ := entry.Data["content"].(string)

	// Emit processing event before starting the HTTP request
	emit(MessageProcessingEvent{
		ConversationID: conversationID,
		UserMessageID:  entry.RowID,
	})

	// Track assistant message state
	var assistantMsgID string
	var model string
	var stopReason string
	messageCreated := false

	// Accumulate content blocks as structured data
	acc := block.NewAccumulator()

	err := h.messages.UploadUserMessage(ctx, entry.RowID, conversationID, content, func(event chat.StreamEvent) error {
		if event.Done {
			h.logger.Debug("stream complete", "assistantMsgID", assistantMsgID, "stopReason", stopReason)
			// Update stop_reason when stream completes (if we have one from the stream)
			if messageCreated && stopReason != "" {
				if err := h.db.Messages().UpdateMeta(ctx, assistantMsgID, model, stopReason); err != nil {
					h.logger.Warn("failed to update assistant message meta", "error", err)
				}
			}
			return nil
		}

		// Capture stop_reason from message_stop event
		if event.Type == block.TypeMessageStop && event.MessageStop != nil {
			stopReason = event.MessageStop.StopReason
			return nil
		}

		// Create assistant message on stream start
		if event.Type == block.TypeMessageStart && event.MessageStart != nil {
			model = event.MessageStart.Model
			var err error
			assistantMsgID, err = h.db.Messages().CreateAssistantMessage(ctx, accountID, conversationID, model)
			if err != nil {
				return fmt.Errorf("create assistant message: %w", err)
			}
			messageCreated = true
			h.logger.Debug("created assistant message", "id", assistantMsgID, "model", model)
			return nil
		}

		// Accumulator handles all content block types
		if acc.Apply(event.Block) && messageCreated {
			if err := h.db.Messages().UpdateContent(ctx, assistantMsgID, acc.JSON()); err != nil {
				h.logger.Warn("failed to update assistant message", "error", err)
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
