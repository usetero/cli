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

func (h *conversationHandler) Handle(ctx context.Context, entry *powersync.CrudEntry) error {
	switch entry.Op {
	case powersync.OpPut:
		return h.handlePut(ctx, entry)
	case powersync.OpPatch:
		// TODO: handle updates if needed
		h.logger.Debug("skipping conversation PATCH", "id", entry.RowID)
		return nil
	case powersync.OpDelete:
		// TODO: handle deletes if needed
		h.logger.Debug("skipping conversation DELETE", "id", entry.RowID)
		return nil
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
