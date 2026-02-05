package upload

// Message Upload Flow
//
// The upload queue handles DURABILITY ONLY. It persists messages to the
// control plane via GraphQL so they survive device loss and sync across devices.
//
// The upload queue does NOT call the Chat API. The TUI calls the Chat API
// directly with the full message history and streams the response.
//
// Flow:
//   1. User submits message in TUI
//   2. TUI writes user message to SQLite (creates ps_crud PUT entry)
//   3. TUI calls Chat API directly, streams response
//   4. TUI writes assistant message to SQLite as deltas arrive
//   5. Upload queue (background) persists messages to GraphQL
//
// For user messages: upload on PUT (initial create)
// For assistant messages: upload when complete (entry contains stop_reason)

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync/db"
	"github.com/usetero/cli/internal/sqlite"
)

// messageHandler persists messages to the control plane via GraphQL.
type messageHandler struct {
	messages api.Messages
	db       sqlite.DB
	scope    log.Scope
}

func newMessageHandler(messages api.Messages, db sqlite.DB, scope log.Scope) *messageHandler {
	return &messageHandler{
		messages: messages,
		db:       db,
		scope:    scope.Child("messages"),
	}
}

// Handle persists a message to the control plane.
func (h *messageHandler) Handle(ctx context.Context, entry *db.CrudEntry, emit Emitter) error {
	switch entry.Op {
	case db.OpPut, db.OpPatch:
		return h.handlePutOrPatch(ctx, entry)
	case db.OpDelete:
		// Messages are deleted via conversation deletion, not individually.
		h.scope.Debug("skipping message DELETE", "id", entry.RowID)
		return nil
	default:
		h.scope.Warn("unknown message op", "op", entry.Op)
		return nil
	}
}

func (h *messageHandler) handlePutOrPatch(ctx context.Context, entry *db.CrudEntry) error {
	msg, err := h.db.Messages().Get(ctx, domain.MessageID(entry.RowID))
	if err != nil {
		// Message may have been deleted, skip
		h.scope.Debug("message not found, skipping", "id", entry.RowID, "error", err)
		return nil
	}

	switch msg.Role {
	case domain.RoleUser:
		// User messages: persist on PUT (initial create)
		if entry.Op == db.OpPut {
			return h.persistUserMessage(ctx, entry, msg)
		}
		h.scope.Debug("skipping user message PATCH", "id", entry.RowID)
		return nil

	case domain.RoleAssistant:
		// Assistant messages: persist when complete (entry sets stop_reason)
		_, hasStopReason := entry.Data["stop_reason"]
		if !hasStopReason {
			h.scope.Debug("skipping assistant message (incomplete)", "id", entry.RowID, "op", entry.Op)
			return nil
		}
		return h.persistAssistantMessage(ctx, msg)

	default:
		h.scope.Warn("unknown message role", "role", msg.Role)
		return nil
	}
}

func (h *messageHandler) persistUserMessage(ctx context.Context, entry *db.CrudEntry, msg *domain.Message) error {
	err := h.messages.CreateMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("persist user message: %w", err)
	}

	h.scope.Debug("persisted user message", "id", entry.RowID)
	return nil
}

func (h *messageHandler) persistAssistantMessage(ctx context.Context, msg *domain.Message) error {
	err := h.messages.CreateMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("persist assistant message: %w", err)
	}

	h.scope.Debug("persisted assistant message", "id", msg.ID)
	return nil
}
