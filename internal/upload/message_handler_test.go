package upload

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/db"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
)

func TestMessageHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("persists user message on PUT", func(t *testing.T) {
		t.Parallel()

		testDB := dbtest.OpenTestDB(t)

		_, err := testDB.Exec(context.Background(),
			`INSERT INTO messages (id, account_id, conversation_id, content, role, created_at)
			 VALUES ('msg-1', 'acc-1', 'conv-1', '[{"type":"text","text":{"content":"Hello"}}]', 'user', '2024-01-01T00:00:00Z')`)
		if err != nil {
			t.Fatalf("insert message: %v", err)
		}

		var captured *domain.Message

		mock := &apitest.MockMessages{
			CreateMessageFunc: func(ctx context.Context, msg *domain.Message) error {
				captured = msg
				return nil
			},
		}

		h := newMessageHandler(mock, testDB, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPut,
			RowID: "msg-1",
			Data: map[string]any{
				"role":            "user",
				"conversation_id": "conv-1",
				"account_id":      "acc-1",
				"content":         `[{"type":"text","text":{"content":"Hello"}}]`,
			},
		}

		err = h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if captured == nil {
			t.Fatal("expected CreateMessage to be called")
		}
		if captured.ID != "msg-1" {
			t.Errorf("id = %q, want %q", captured.ID, "msg-1")
		}
		if captured.ConversationID != "conv-1" {
			t.Errorf("conversationID = %q, want %q", captured.ConversationID, "conv-1")
		}
		if captured.Role != domain.RoleUser {
			t.Errorf("role = %q, want %q", captured.Role, domain.RoleUser)
		}
	})

	t.Run("skips user message on PATCH", func(t *testing.T) {
		t.Parallel()

		testDB := dbtest.OpenTestDB(t)

		_, err := testDB.Exec(context.Background(),
			`INSERT INTO messages (id, account_id, conversation_id, content, role, created_at)
			 VALUES ('msg-1', 'acc-1', 'conv-1', '[{"type":"text","text":{"content":"Hello"}}]', 'user', '2024-01-01T00:00:00Z')`)
		if err != nil {
			t.Fatalf("insert message: %v", err)
		}

		called := false
		mock := &apitest.MockMessages{
			CreateMessageFunc: func(ctx context.Context, msg *domain.Message) error {
				called = true
				return nil
			},
		}

		h := newMessageHandler(mock, testDB, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPatch,
			RowID: "msg-1",
			Data: map[string]any{
				"content": `[{"type":"text","text":{"content":"Hello updated"}}]`,
			},
		}

		err = h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
		if called {
			t.Error("expected CreateMessage to not be called on PATCH for user message")
		}
	})

	t.Run("persists assistant message when stop_reason is set", func(t *testing.T) {
		t.Parallel()

		testDB := dbtest.OpenTestDB(t)

		_, err := testDB.Exec(context.Background(),
			`INSERT INTO messages (id, account_id, conversation_id, content, role, model, stop_reason, created_at)
			 VALUES ('msg-1', 'acc-1', 'conv-1', '[{"type":"text","text":{"content":"Hi"}}]', 'assistant', 'claude-3', 'end_turn', '2024-01-01T00:00:00Z')`)
		if err != nil {
			t.Fatalf("insert message: %v", err)
		}

		var captured *domain.Message

		mock := &apitest.MockMessages{
			CreateMessageFunc: func(ctx context.Context, msg *domain.Message) error {
				captured = msg
				return nil
			},
		}

		h := newMessageHandler(mock, testDB, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPatch,
			RowID: "msg-1",
			Data: map[string]any{
				"stop_reason": "end_turn",
			},
		}

		err = h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if captured == nil {
			t.Fatal("expected CreateMessage to be called")
		}
		if captured.ID != "msg-1" {
			t.Errorf("id = %q, want %q", captured.ID, "msg-1")
		}
		if captured.Model != "claude-3" {
			t.Errorf("model = %q, want %q", captured.Model, "claude-3")
		}
		if captured.StopReason != "end_turn" {
			t.Errorf("stopReason = %q, want %q", captured.StopReason, "end_turn")
		}
	})

	t.Run("skips assistant message without stop_reason", func(t *testing.T) {
		t.Parallel()

		testDB := dbtest.OpenTestDB(t)

		_, err := testDB.Exec(context.Background(),
			`INSERT INTO messages (id, account_id, conversation_id, content, role, model, created_at)
			 VALUES ('msg-1', 'acc-1', 'conv-1', '[]', 'assistant', 'claude-3', '2024-01-01T00:00:00Z')`)
		if err != nil {
			t.Fatalf("insert message: %v", err)
		}

		called := false
		mock := &apitest.MockMessages{
			CreateMessageFunc: func(ctx context.Context, msg *domain.Message) error {
				called = true
				return nil
			},
		}

		h := newMessageHandler(mock, testDB, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPatch,
			RowID: "msg-1",
			Data: map[string]any{
				"content": `[{"type":"text","text":{"content":"partial"}}]`,
			},
		}

		err = h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
		if called {
			t.Error("expected CreateMessage to not be called without stop_reason")
		}
	})

	t.Run("returns error when persist fails", func(t *testing.T) {
		t.Parallel()

		testDB := dbtest.OpenTestDB(t)

		_, err := testDB.Exec(context.Background(),
			`INSERT INTO messages (id, account_id, conversation_id, content, role, created_at)
			 VALUES ('msg-1', 'acc-1', 'conv-1', '[{"type":"text","text":{"content":"Hello"}}]', 'user', '2024-01-01T00:00:00Z')`)
		if err != nil {
			t.Fatalf("insert message: %v", err)
		}

		mock := &apitest.MockMessages{
			CreateMessageFunc: func(ctx context.Context, msg *domain.Message) error {
				return errors.New("network error")
			},
		}

		h := newMessageHandler(mock, testDB, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPut,
			RowID: "msg-1",
			Data: map[string]any{
				"role":            "user",
				"conversation_id": "conv-1",
				"content":         `[{"type":"text","text":{"content":"Hello"}}]`,
			},
		}

		err = h.Handle(context.Background(), entry, noopEmitter())
		if err == nil {
			t.Error("Handle() expected error, got nil")
		}
	})

	t.Run("skips DELETE", func(t *testing.T) {
		t.Parallel()

		testDB := dbtest.OpenTestDB(t)
		h := newMessageHandler(apitest.NewMockMessages(), testDB, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpDelete,
			RowID: "msg-1",
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Errorf("Handle() error = %v, want nil", err)
		}
	})

	t.Run("skips unknown role", func(t *testing.T) {
		t.Parallel()

		testDB := dbtest.OpenTestDB(t)

		_, err := testDB.Exec(context.Background(),
			`INSERT INTO messages (id, account_id, conversation_id, content, role, created_at)
			 VALUES ('msg-1', 'acc-1', 'conv-1', '[]', 'system', '2024-01-01T00:00:00Z')`)
		if err != nil {
			t.Fatalf("insert message: %v", err)
		}

		h := newMessageHandler(apitest.NewMockMessages(), testDB, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Op:    db.OpPut,
			RowID: "msg-1",
			Data: map[string]any{
				"role": "system",
			},
		}

		err = h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Errorf("Handle() error = %v, want nil", err)
		}
	})
}
