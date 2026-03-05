package upload

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync/db"
)

// Idempotent operation handling:
// - DELETE returning "not found" is success (resource already gone)
// - CREATE returning "already exists" is success (resource already there)

// conversationHandler handles uploading conversations to the GraphQL API.
type conversationHandler struct {
	conversations graphql.Conversations
	scope         log.Scope
}

func newConversationHandler(conversations graphql.Conversations, scope log.Scope) *conversationHandler {
	return &conversationHandler{
		conversations: conversations,
		scope:         scope.Child("conversations"),
	}
}

func (h *conversationHandler) Handle(ctx context.Context, entry *db.CrudEntry, emit Emitter) error {
	// Conversation handler doesn't emit events currently
	_ = emit

	switch entry.Op {
	case db.OpPut:
		return h.handlePut(ctx, entry)
	case db.OpPatch:
		return h.handlePatch(ctx, entry)
	case db.OpDelete:
		return h.handleDelete(ctx, entry)
	default:
		h.scope.Warn("unknown conversation op", "op", entry.Op)
		return nil
	}
}

func (h *conversationHandler) handlePut(ctx context.Context, entry *db.CrudEntry) error {
	workspaceID, _ := entry.Data["workspace_id"].(string)
	title, _ := entry.Data["title"].(string)

	id, err := uuid.Parse(entry.RowID)
	if err != nil {
		return fmt.Errorf("invalid conversation ID %q: %w", entry.RowID, err)
	}

	_, err = h.conversations.Create(ctx, graphql.CreateConversationInput{
		ID:          id,
		WorkspaceID: domain.WorkspaceID(workspaceID),
		Title:       title,
	})
	if err != nil {
		// Already exists is fine - the conversation is there, which is what we wanted
		if errors.Is(err, graphql.ErrAlreadyExists) {
			h.scope.Debug("conversation already exists, skipping", "id", entry.RowID)
			return nil
		}
		return fmt.Errorf("create conversation: %w", err)
	}

	h.scope.Debug("uploaded conversation", "id", entry.RowID)
	return nil
}

func (h *conversationHandler) handlePatch(ctx context.Context, entry *db.CrudEntry) error {
	var input graphql.UpdateConversationInput

	if titleVal, ok := entry.Data["title"]; ok {
		title, _ := titleVal.(string)
		input.Title = &title
	}

	_, err := h.conversations.Update(ctx, domain.ConversationID(entry.RowID), input)
	if err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}

	h.scope.Debug("updated conversation", "id", entry.RowID)
	return nil
}

func (h *conversationHandler) handleDelete(ctx context.Context, entry *db.CrudEntry) error {
	err := h.conversations.Delete(ctx, domain.ConversationID(entry.RowID))
	if err != nil {
		// Not found is fine - the conversation is gone, which is what we wanted
		if errors.Is(err, graphql.ErrNotFound) {
			h.scope.Debug("conversation already deleted, skipping", "id", entry.RowID)
			return nil
		}
		return fmt.Errorf("delete conversation: %w", err)
	}

	h.scope.Debug("deleted conversation", "id", entry.RowID)
	return nil
}
