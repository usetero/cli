package upload

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
)

func noopEmitter() Emitter {
	return func(Event) {}
}

func TestConversationHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("PUT creates conversation", func(t *testing.T) {
		t.Parallel()

		testID := uuid.New()
		var calledWith struct {
			id          uuid.UUID
			workspaceID string
			title       string
		}

		mock := &apitest.MockConversations{
			CreateFunc: func(ctx context.Context, id uuid.UUID, workspaceID, title string) (*api.Conversation, error) {
				calledWith.id = id
				calledWith.workspaceID = workspaceID
				calledWith.title = title
				return &api.Conversation{ID: id.String()}, nil
			},
		}

		h := newConversationHandler(mock, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPut,
			RowID: testID.String(),
			Data: map[string]any{
				"workspace_id": "ws-1",
				"title":        "Test Conversation",
			},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if calledWith.id != testID {
			t.Errorf("Create called with id = %v, want %v", calledWith.id, testID)
		}
		if calledWith.workspaceID != "ws-1" {
			t.Errorf("Create called with workspaceID = %q, want %q", calledWith.workspaceID, "ws-1")
		}
		if calledWith.title != "Test Conversation" {
			t.Errorf("Create called with title = %q, want %q", calledWith.title, "Test Conversation")
		}
	})

	t.Run("PUT returns error on failure", func(t *testing.T) {
		t.Parallel()

		mock := &apitest.MockConversations{
			CreateFunc: func(ctx context.Context, id uuid.UUID, workspaceID, title string) (*api.Conversation, error) {
				return nil, errors.New("network error")
			},
		}

		h := newConversationHandler(mock, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPut,
			RowID: uuid.New().String(),
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err == nil {
			t.Error("Handle() expected error, got nil")
		}
	})

	t.Run("PATCH updates conversation", func(t *testing.T) {
		t.Parallel()

		var calledWith struct {
			id    string
			title string
		}

		mock := &apitest.MockConversations{
			UpdateFunc: func(ctx context.Context, id, title string) (*api.Conversation, error) {
				calledWith.id = id
				calledWith.title = title
				return &api.Conversation{ID: id, Title: title}, nil
			},
		}

		h := newConversationHandler(mock, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPatch,
			RowID: "conv-1",
			Data: map[string]any{
				"title": "Updated Title",
			},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if calledWith.id != "conv-1" {
			t.Errorf("Update called with id = %q, want %q", calledWith.id, "conv-1")
		}
		if calledWith.title != "Updated Title" {
			t.Errorf("Update called with title = %q, want %q", calledWith.title, "Updated Title")
		}
	})

	t.Run("PATCH returns error on failure", func(t *testing.T) {
		t.Parallel()

		mock := &apitest.MockConversations{
			UpdateFunc: func(ctx context.Context, id, title string) (*api.Conversation, error) {
				return nil, errors.New("network error")
			},
		}

		h := newConversationHandler(mock, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPatch,
			RowID: "conv-1",
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err == nil {
			t.Error("Handle() expected error, got nil")
		}
	})

	t.Run("DELETE deletes conversation", func(t *testing.T) {
		t.Parallel()

		var deletedID string

		mock := &apitest.MockConversations{
			DeleteFunc: func(ctx context.Context, id string) error {
				deletedID = id
				return nil
			},
		}

		h := newConversationHandler(mock, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpDelete,
			RowID: "conv-1",
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if deletedID != "conv-1" {
			t.Errorf("Delete called with id = %q, want %q", deletedID, "conv-1")
		}
	})

	t.Run("DELETE returns error on failure", func(t *testing.T) {
		t.Parallel()

		mock := &apitest.MockConversations{
			DeleteFunc: func(ctx context.Context, id string) error {
				return errors.New("network error")
			},
		}

		h := newConversationHandler(mock, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpDelete,
			RowID: "conv-1",
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err == nil {
			t.Error("Handle() expected error, got nil")
		}
	})

	t.Run("DELETE succeeds when resource not found", func(t *testing.T) {
		t.Parallel()

		mock := &apitest.MockConversations{
			DeleteFunc: func(ctx context.Context, id string) error {
				return errors.New("conversation not found")
			},
		}

		h := newConversationHandler(mock, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpDelete,
			RowID: "conv-1",
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Errorf("Handle() error = %v, want nil for not found (idempotent delete)", err)
		}
	})

	t.Run("PUT succeeds when resource already exists", func(t *testing.T) {
		t.Parallel()

		mock := &apitest.MockConversations{
			CreateFunc: func(ctx context.Context, id uuid.UUID, workspaceID, title string) (*api.Conversation, error) {
				return nil, errors.New("conversation already exists")
			},
		}

		h := newConversationHandler(mock, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPut,
			RowID: uuid.New().String(),
			Data: map[string]any{
				"workspace_id": "ws-1",
				"title":        "Test",
			},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Errorf("Handle() error = %v, want nil for already exists (idempotent create)", err)
		}
	})

	t.Run("unknown op returns nil", func(t *testing.T) {
		t.Parallel()

		h := newConversationHandler(&apitest.MockConversations{}, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    "UNKNOWN",
			RowID: "conv-1",
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Errorf("Handle() error = %v, want nil for unknown op", err)
		}
	})
}
