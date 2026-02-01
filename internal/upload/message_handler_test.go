package upload

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/chat/chattest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
)

func TestMessageHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("uploads user message with correct parameters", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		var captured struct {
			messageID      string
			conversationID string
			content        string
		}

		mock := &chattest.MockMessages{
			UploadUserMessageFunc: func(ctx context.Context, messageID, conversationID, content string, handler chat.StreamHandler) error {
				captured.messageID = messageID
				captured.conversationID = conversationID
				captured.content = content
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

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if captured.messageID != "msg-1" {
			t.Errorf("messageID = %q, want %q", captured.messageID, "msg-1")
		}
		if captured.conversationID != "conv-1" {
			t.Errorf("conversationID = %q, want %q", captured.conversationID, "conv-1")
		}
		if captured.content != "Hello" {
			t.Errorf("content = %q, want %q", captured.content, "Hello")
		}
	})

	t.Run("creates assistant message and accumulates streaming deltas", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		mock := &chattest.MockMessages{
			UploadUserMessageFunc: func(ctx context.Context, messageID, conversationID, content string, handler chat.StreamHandler) error {
				events := []chat.StreamEvent{
					{Block: block.Block{Type: block.TypeMessageStart}},
					{Block: block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "Hello"}}},
					{Block: block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: " world"}}},
					{Block: block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "!"}}},
					{Done: true},
				}
				for _, e := range events {
					if err := handler(e); err != nil {
						return err
					}
				}
				return nil
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

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		// Query assistant message from SQLite
		rows, err := db.Query(context.Background(), "SELECT content FROM messages WHERE role = 'assistant'")
		if err != nil {
			t.Fatalf("Query error = %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Fatal("expected assistant message in database")
		}

		var contentJSON string
		if err := rows.Scan(&contentJSON); err != nil {
			t.Fatalf("Scan error = %v", err)
		}

		// Parse and verify content blocks
		var blocks []block.Block
		if err := json.Unmarshal([]byte(contentJSON), &blocks); err != nil {
			t.Fatalf("content is not valid JSON: %v", err)
		}

		if len(blocks) != 1 {
			t.Fatalf("got %d blocks, want 1", len(blocks))
		}
		if blocks[0].Type != block.TypeText {
			t.Errorf("block type = %q, want %q", blocks[0].Type, block.TypeText)
		}
		if blocks[0].Text.Content != "Hello world!" {
			t.Errorf("text content = %q, want %q", blocks[0].Text.Content, "Hello world!")
		}
	})

	t.Run("accumulates thinking then text as separate blocks", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		mock := &chattest.MockMessages{
			UploadUserMessageFunc: func(ctx context.Context, messageID, conversationID, content string, handler chat.StreamHandler) error {
				events := []chat.StreamEvent{
					{Block: block.Block{Type: block.TypeMessageStart}},
					{Block: block.Block{Type: block.TypeThinkingDelta, Thinking: &block.Thinking{Content: "Let me think"}}},
					{Block: block.Block{Type: block.TypeThinkingDelta, Thinking: &block.Thinking{Content: "..."}}},
					{Block: block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "Here's my answer"}}},
					{Done: true},
				}
				for _, e := range events {
					if err := handler(e); err != nil {
						return err
					}
				}
				return nil
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
				"content":         "Question",
			},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		rows, err := db.Query(context.Background(), "SELECT content FROM messages WHERE role = 'assistant'")
		if err != nil {
			t.Fatalf("Query error = %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Fatal("expected assistant message in database")
		}

		var contentJSON string
		if err := rows.Scan(&contentJSON); err != nil {
			t.Fatalf("Scan error = %v", err)
		}

		var blocks []block.Block
		if err := json.Unmarshal([]byte(contentJSON), &blocks); err != nil {
			t.Fatalf("content is not valid JSON: %v", err)
		}

		if len(blocks) != 2 {
			t.Fatalf("got %d blocks, want 2", len(blocks))
		}
		if blocks[0].Type != block.TypeThinking {
			t.Errorf("first block type = %q, want %q", blocks[0].Type, block.TypeThinking)
		}
		if blocks[0].Thinking.Content != "Let me think..." {
			t.Errorf("thinking content = %q, want %q", blocks[0].Thinking.Content, "Let me think...")
		}
		if blocks[1].Type != block.TypeText {
			t.Errorf("second block type = %q, want %q", blocks[1].Type, block.TypeText)
		}
	})

	t.Run("returns error when upload fails", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

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

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err == nil {
			t.Error("Handle() expected error, got nil")
		}
	})

	t.Run("uploads assistant message for durability", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		var captured struct {
			messageID  string
			content    string
			model      string
			stopReason string
		}

		mock := &chattest.MockMessages{
			UploadAssistantMessageFunc: func(ctx context.Context, messageID, conversationID, content, model, stopReason string) error {
				captured.messageID = messageID
				captured.content = content
				captured.model = model
				captured.stopReason = stopReason
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
				"content":         `[{"type":"text","text":{"content":"Hi there!"}}]`,
				"model":           "claude-3",
				"stop_reason":     "end_turn",
			},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if captured.messageID != "msg-2" {
			t.Errorf("messageID = %q, want %q", captured.messageID, "msg-2")
		}
		if captured.model != "claude-3" {
			t.Errorf("model = %q, want %q", captured.model, "claude-3")
		}
	})

	t.Run("PATCH skips upload", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		h := newMessageHandler(&chattest.MockMessages{}, db, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPatch,
			RowID: "msg-1",
			Data:  map[string]any{},
		}

		if err := h.Handle(context.Background(), entry, noopEmitter()); err != nil {
			t.Errorf("Handle() error = %v, want nil", err)
		}
	})

	t.Run("DELETE skips upload", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		h := newMessageHandler(&chattest.MockMessages{}, db, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpDelete,
			RowID: "msg-1",
			Data:  map[string]any{},
		}

		if err := h.Handle(context.Background(), entry, noopEmitter()); err != nil {
			t.Errorf("Handle() error = %v, want nil", err)
		}
	})

	t.Run("unknown role skips upload", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		h := newMessageHandler(&chattest.MockMessages{}, db, logtest.New(t))

		entry := &powersync.CrudEntry{
			Op:    powersync.OpPut,
			RowID: "msg-1",
			Data: map[string]any{
				"role": "system",
			},
		}

		if err := h.Handle(context.Background(), entry, noopEmitter()); err != nil {
			t.Errorf("Handle() error = %v, want nil", err)
		}
	})
}
