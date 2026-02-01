package upload

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
)

// conversationHandler handles uploading conversations to the GraphQL API.
type conversationHandler struct {
	conversations api.Conversations
	logger        log.Logger
}

func newConversationHandler(conversations api.Conversations, logger log.Logger) *conversationHandler {
	return &conversationHandler{
		conversations: conversations,
		logger:        logger,
	}
}

func (h *conversationHandler) Handle(ctx context.Context, entry *powersync.CrudEntry, emit Emitter) error {
	// Conversation handler doesn't emit events currently
	_ = emit

	switch entry.Op {
	case powersync.OpPut:
		return h.handlePut(ctx, entry)
	case powersync.OpPatch:
		return h.handlePatch(ctx, entry)
	case powersync.OpDelete:
		return h.handleDelete(ctx, entry)
	default:
		h.logger.Warn("unknown conversation op", "op", entry.Op)
		return nil
	}
}

func (h *conversationHandler) handlePut(ctx context.Context, entry *powersync.CrudEntry) error {
	workspaceID, _ := entry.Data["workspace_id"].(string)
	title, _ := entry.Data["title"].(string)

	_, err := h.conversations.Create(ctx, workspaceID, title)
	if err != nil {
		return fmt.Errorf("create conversation: %w", err)
	}

	h.logger.Debug("uploaded conversation", "id", entry.RowID)
	return nil
}

func (h *conversationHandler) handlePatch(ctx context.Context, entry *powersync.CrudEntry) error {
	title, _ := entry.Data["title"].(string)

	_, err := h.conversations.Update(ctx, entry.RowID, title)
	if err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}

	h.logger.Debug("updated conversation", "id", entry.RowID)
	return nil
}

func (h *conversationHandler) handleDelete(ctx context.Context, entry *powersync.CrudEntry) error {
	err := h.conversations.Delete(ctx, entry.RowID)
	if err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}

	h.logger.Debug("deleted conversation", "id", entry.RowID)
	return nil
}
