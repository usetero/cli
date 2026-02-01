package upload

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/chattest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
)

func TestMessageHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("PUT user message uploads and streams response", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDBWithSchema(t)

		var calledWith struct {
			messageID      string
			conversationID string
			content        string
		}

		mock := &chattest.MockMessages{
			UploadUserMessageFunc: func(ctx context.Context, messageID, conversationID, content string, handler chat.StreamHandler) error {
				calledWith.messageID = messageID
				calledWith.conversationID = conversationID
				calledWith.content = content
				// Simulate stream completion
				return handler(chat.StreamEvent{Done: true})
			},
		}

		h := newMessageHandler(mock, db, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPut,
			RowID: "msg-1",
			Data: map[string]any{
				"role":            "user",
				"conversation_id": "conv-1",
				"account_id":      "acc-1",
				"content":         "Hello",
			},
		}

		err := h.Handle(context.Background(), entry)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if calledWith.messageID != "msg-1" {
			t.Errorf("UploadUserMessage called with messageID = %q, want %q", calledWith.messageID, "msg-1")
		}
		if calledWith.conversationID != "conv-1" {
			t.Errorf("UploadUserMessage called with conversationID = %q, want %q", calledWith.conversationID, "conv-1")
		}
		if calledWith.content != "Hello" {
			t.Errorf("UploadUserMessage called with content = %q, want %q", calledWith.content, "Hello")
		}
	})

	t.Run("PUT user message creates assistant message and updates with deltas", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDBWithSchema(t)

		mock := &chattest.MockMessages{
			UploadUserMessageFunc: func(ctx context.Context, messageID, conversationID, content string, handler chat.StreamHandler) error {
				// Simulate streaming: message_start, then multiple deltas, then done
				if err := handler(chat.StreamEvent{Type: "message_start"}); err != nil {
					return err
				}
				if err := handler(chat.StreamEvent{Type: "content_block_delta", Text: &chat.TextDelta{Content: "Hello"}}); err != nil {
					return err
				}
				if err := handler(chat.StreamEvent{Type: "content_block_delta", Text: &chat.TextDelta{Content: " world"}}); err != nil {
					return err
				}
				if err := handler(chat.StreamEvent{Type: "content_block_delta", Text: &chat.TextDelta{Content: "!"}}); err != nil {
					return err
				}
				return handler(chat.StreamEvent{Done: true})
			},
		}

		h := newMessageHandler(mock, db, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPut,
			RowID: "msg-1",
			Data: map[string]any{
				"role":            "user",
				"conversation_id": "conv-1",
				"account_id":      "acc-1",
				"content":         "Hi",
			},
		}

		err := h.Handle(context.Background(), entry)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		// Verify assistant message was created and updated in SQLite
		rows, err := db.Query(context.Background(), "SELECT content FROM messages WHERE role = 'assistant'")
		if err != nil {
			t.Fatalf("Query error = %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Fatal("expected assistant message in database")
		}

		var content string
		if err := rows.Scan(&content); err != nil {
			t.Fatalf("Scan error = %v", err)
		}

		if content != "Hello world!" {
			t.Errorf("assistant message content = %q, want %q", content, "Hello world!")
		}
	})

	t.Run("PUT user message returns error on failure", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDBWithSchema(t)

		mock := &chattest.MockMessages{
			UploadUserMessageFunc: func(ctx context.Context, messageID, conversationID, content string, handler chat.StreamHandler) error {
				return errors.New("network error")
			},
		}

		h := newMessageHandler(mock, db, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPut,
			RowID: "msg-1",
			Data: map[string]any{
				"role":            "user",
				"conversation_id": "conv-1",
				"content":         "Hello",
			},
		}

		err := h.Handle(context.Background(), entry)
		if err == nil {
			t.Error("Handle() expected error, got nil")
		}
	})

	t.Run("PUT assistant message uploads for durability", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDBWithSchema(t)

		var calledWith struct {
			messageID      string
			conversationID string
			content        string
			model          string
			stopReason     string
		}

		mock := &chattest.MockMessages{
			UploadAssistantMessageFunc: func(ctx context.Context, messageID, conversationID, content, model, stopReason string) error {
				calledWith.messageID = messageID
				calledWith.conversationID = conversationID
				calledWith.content = content
				calledWith.model = model
				calledWith.stopReason = stopReason
				return nil
			},
		}

		h := newMessageHandler(mock, db, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPut,
			RowID: "msg-2",
			Data: map[string]any{
				"role":            "assistant",
				"conversation_id": "conv-1",
				"content":         "Hi there!",
				"model":           "claude-3",
				"stop_reason":     "end_turn",
			},
		}

		err := h.Handle(context.Background(), entry)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if calledWith.messageID != "msg-2" {
			t.Errorf("UploadAssistantMessage called with messageID = %q, want %q", calledWith.messageID, "msg-2")
		}
		if calledWith.content != "Hi there!" {
			t.Errorf("UploadAssistantMessage called with content = %q, want %q", calledWith.content, "Hi there!")
		}
	})

	t.Run("PATCH skips upload and returns nil", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDBWithSchema(t)
		h := newMessageHandler(&chattest.MockMessages{}, db, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPatch,
			RowID: "msg-1",
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry)
		if err != nil {
			t.Errorf("Handle() error = %v, want nil for PATCH", err)
		}
	})

	t.Run("DELETE skips upload and returns nil", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDBWithSchema(t)
		h := newMessageHandler(&chattest.MockMessages{}, db, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpDelete,
			RowID: "msg-1",
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry)
		if err != nil {
			t.Errorf("Handle() error = %v, want nil for DELETE", err)
		}
	})

	t.Run("unknown role returns nil", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDBWithSchema(t)
		h := newMessageHandler(&chattest.MockMessages{}, db, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPut,
			RowID: "msg-1",
			Data: map[string]any{
				"role": "system",
			},
		}

		err := h.Handle(context.Background(), entry)
		if err != nil {
			t.Errorf("Handle() error = %v, want nil for unknown role", err)
		}
	})
}
