package upload

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/apitest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/db"
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
			accountID   domain.AccountID
			workspaceID domain.WorkspaceID
			title       string
		}

		mock := &apitest.MockConversations{
			CreateFunc: func(ctx context.Context, input graphql.CreateConversationInput) (*domain.Conversation, error) {
				calledWith.id = input.ID
				calledWith.accountID = input.AccountID
				calledWith.workspaceID = input.WorkspaceID
				calledWith.title = input.Title
				return &domain.Conversation{ID: domain.ConversationID(input.ID.String())}, nil
			},
		}

		h := newConversationHandler(mock, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPut,
			RowID: testID.String(),
			Data: map[string]any{
				"account_id":   "acc-1",
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
		if calledWith.accountID != "acc-1" {
			t.Errorf("Create called with accountID = %q, want %q", calledWith.accountID, "acc-1")
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
			CreateFunc: func(ctx context.Context, input graphql.CreateConversationInput) (*domain.Conversation, error) {
				return nil, errors.New("network error")
			},
		}

		h := newConversationHandler(mock, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPut,
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
			id    domain.ConversationID
			title string
		}

		mock := &apitest.MockConversations{
			UpdateFunc: func(ctx context.Context, id domain.ConversationID, input graphql.UpdateConversationInput) (*domain.Conversation, error) {
				calledWith.id = id
				if input.Title != nil {
					calledWith.title = *input.Title
				}
				return &domain.Conversation{ID: id, Title: calledWith.title}, nil
			},
		}

		h := newConversationHandler(mock, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPatch,
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
			UpdateFunc: func(ctx context.Context, id domain.ConversationID, input graphql.UpdateConversationInput) (*domain.Conversation, error) {
				return nil, errors.New("network error")
			},
		}

		h := newConversationHandler(mock, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPatch,
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

		var deletedID domain.ConversationID

		mock := &apitest.MockConversations{
			DeleteFunc: func(ctx context.Context, id domain.ConversationID) error {
				deletedID = id
				return nil
			},
		}

		h := newConversationHandler(mock, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpDelete,
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
			DeleteFunc: func(ctx context.Context, id domain.ConversationID) error {
				return errors.New("network error")
			},
		}

		h := newConversationHandler(mock, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpDelete,
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
			DeleteFunc: func(ctx context.Context, id domain.ConversationID) error {
				// Service layer returns wrapped ErrNotFound
				return fmt.Errorf("delete conversation %s: %w", id, graphql.ErrNotFound)
			},
		}

		h := newConversationHandler(mock, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpDelete,
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
			CreateFunc: func(ctx context.Context, input graphql.CreateConversationInput) (*domain.Conversation, error) {
				// Service layer returns wrapped ErrAlreadyExists
				return nil, fmt.Errorf("create conversation %s: %w", input.ID, graphql.ErrAlreadyExists)
			},
		}

		h := newConversationHandler(mock, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPut,
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

		h := newConversationHandler(apitest.NewMockConversations(), logtest.NewScope(t))

		entry := &db.CrudEntry{
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
