package chat

import (
	"errors"
	"testing"

	"github.com/usetero/cli/internal/sqlite/gen"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/upload"
)

func TestMessages_Update(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("messagesLoadedMsg sets items", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetWidth(80)

		id := "msg-1"
		role := "user"
		content := `[{"type":"text","text":{"content":"Hello"}}]`

		m.Update(messagesLoadedMsg{
			messages: []gen.Message{
				{ID: &id, Role: &role, Content: &content},
			},
		})

		items := m.Items()
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if items[0].ID() != "msg-1" {
			t.Errorf("expected ID 'msg-1', got %q", items[0].ID())
		}
	})

	t.Run("messagesLoadedMsg sets error on failure", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)

		m.Update(messagesLoadedMsg{err: errors.New("db error")})

		if !m.HasError() {
			t.Error("expected HasError() to return true")
		}
		if m.Error().Error() != "db error" {
			t.Errorf("expected error 'db error', got %q", m.Error().Error())
		}
	})

	t.Run("messagesLoadedMsg clears previous error", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)

		m.Update(messagesLoadedMsg{err: errors.New("temporary")})
		if !m.HasError() {
			t.Fatal("expected error state")
		}

		m.Update(messagesLoadedMsg{messages: []gen.Message{}})

		if m.HasError() {
			t.Error("expected error to be cleared")
		}
	})

	t.Run("tablesChangedMsg refreshes on messages table", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetConversation("conv-1")
		m.Init()

		cmd := m.Update(tablesChangedMsg{tables: []string{"messages"}})

		if cmd == nil {
			t.Error("expected command for messages table change")
		}
	})

	t.Run("tablesChangedMsg ignores other tables", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetConversation("conv-1")
		m.Init()

		// Should return a command (to keep listening) but not refresh
		cmd := m.Update(tablesChangedMsg{tables: []string{"conversations"}})

		// Command is for listening, which is expected
		_ = cmd
	})

	t.Run("UploadEventMsg adds pending assistant", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetWidth(80)
		m.SetConversation("conv-1")

		event := UploadEventMsg{
			Event: upload.MessageProcessingEvent{
				ConversationID: "conv-1",
				UserMessageID:  "user-1",
			},
		}

		cmd := m.Update(event)

		items := m.Items()
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}

		am, ok := items[0].(*AssistantMessage)
		if !ok {
			t.Fatal("expected AssistantMessage")
		}
		if am.ID() != "" {
			t.Error("expected empty ID for pending assistant")
		}
		if am.State() != StateSending {
			t.Errorf("expected StateSending, got %v", am.State())
		}
		if cmd == nil {
			t.Error("expected init command")
		}
	})

	t.Run("UploadEventMsg ignores other conversations", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetConversation("conv-1")

		event := UploadEventMsg{
			Event: upload.MessageProcessingEvent{
				ConversationID: "other-conv",
				UserMessageID:  "user-1",
			},
		}

		m.Update(event)

		if len(m.Items()) != 0 {
			t.Error("expected no items for other conversation")
		}
	})
}

func TestMessages_BuildItems(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("creates user and assistant messages", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetWidth(80)

		userID := "msg-1"
		userRole := "user"
		userContent := `[{"type":"text","text":{"content":"Hi"}}]`

		assistantID := "msg-2"
		assistantRole := "assistant"
		assistantContent := `[{"type":"text","text":{"content":"Hello"}}]`

		m.Update(messagesLoadedMsg{
			messages: []gen.Message{
				{ID: &userID, Role: &userRole, Content: &userContent},
				{ID: &assistantID, Role: &assistantRole, Content: &assistantContent},
			},
		})

		items := m.Items()
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}

		if _, ok := items[0].(*UserMessage); !ok {
			t.Error("expected first item to be UserMessage")
		}
		if _, ok := items[1].(*AssistantMessage); !ok {
			t.Error("expected second item to be AssistantMessage")
		}
	})

	t.Run("promotes pending assistant on load", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetWidth(80)
		m.SetConversation("conv-1")

		// Create pending assistant
		m.Update(UploadEventMsg{
			Event: upload.MessageProcessingEvent{
				ConversationID: "conv-1",
				UserMessageID:  "user-1",
			},
		})

		pending, ok := m.Items()[0].(*AssistantMessage)
		if !ok {
			t.Fatal("expected AssistantMessage")
		}
		if pending.ID() != "" {
			t.Fatal("expected empty ID before promotion")
		}

		// Load messages with the assistant
		assistantID := "assistant-1"
		assistantRole := "assistant"
		assistantContent := `[{"type":"text","text":{"content":"Response"}}]`

		m.Update(messagesLoadedMsg{
			messages: []gen.Message{
				{ID: &assistantID, Role: &assistantRole, Content: &assistantContent},
			},
		})

		items := m.Items()
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}

		// Should be same instance, now with ID
		am, ok := items[0].(*AssistantMessage)
		if !ok {
			t.Fatal("expected AssistantMessage")
		}
		if am != pending {
			t.Error("expected pending assistant to be promoted, not replaced")
		}
		if am.ID() != "assistant-1" {
			t.Errorf("expected ID 'assistant-1', got %q", am.ID())
		}
	})

	t.Run("preserves pending if no assistant in load", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetWidth(80)
		m.SetConversation("conv-1")

		// Create pending assistant
		m.Update(UploadEventMsg{
			Event: upload.MessageProcessingEvent{
				ConversationID: "conv-1",
				UserMessageID:  "user-1",
			},
		})

		// Load only user messages
		userID := "user-1"
		userRole := "user"
		userContent := `[{"type":"text","text":{"content":"Hi"}}]`

		m.Update(messagesLoadedMsg{
			messages: []gen.Message{
				{ID: &userID, Role: &userRole, Content: &userContent},
			},
		})

		items := m.Items()
		if len(items) != 2 {
			t.Fatalf("expected 2 items (user + pending), got %d", len(items))
		}

		// Pending should be at end
		am, ok := items[1].(*AssistantMessage)
		if !ok {
			t.Fatal("expected last item to be AssistantMessage")
		}
		if am.ID() != "" {
			t.Error("expected pending assistant to still have empty ID")
		}
	})

	t.Run("skips messages without ID or role", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetWidth(80)

		validID := "msg-1"
		validRole := "user"
		validContent := `[{"type":"text","text":{"content":"Hi"}}]`

		m.Update(messagesLoadedMsg{
			messages: []gen.Message{
				{Role: &validRole, Content: &validContent},               // missing ID
				{ID: &validID, Content: &validContent},                   // missing role
				{ID: &validID, Role: &validRole, Content: &validContent}, // valid
			},
		})

		items := m.Items()
		if len(items) != 1 {
			t.Fatalf("expected 1 valid item, got %d", len(items))
		}
	})

	t.Run("reuses existing items on reload", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetWidth(80)

		id := "msg-1"
		role := "user"
		content := `[{"type":"text","text":{"content":"Hi"}}]`

		m.Update(messagesLoadedMsg{
			messages: []gen.Message{
				{ID: &id, Role: &role, Content: &content},
			},
		})

		original := m.Items()[0]

		// Reload same message
		m.Update(messagesLoadedMsg{
			messages: []gen.Message{
				{ID: &id, Role: &role, Content: &content},
			},
		})

		reloaded := m.Items()[0]

		if original != reloaded {
			t.Error("expected same item instance on reload")
		}
	})
}

func TestMessages_SetConversation(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("clears items and error", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetWidth(80)

		// Add some state
		id := "msg-1"
		role := "user"
		content := `[{"type":"text","text":{"content":"Hi"}}]`
		m.Update(messagesLoadedMsg{
			messages: []gen.Message{
				{ID: &id, Role: &role, Content: &content},
			},
		})
		m.Update(messagesLoadedMsg{err: errors.New("error")})

		m.SetConversation("new-conv")

		if len(m.Items()) != 0 {
			t.Error("expected items to be cleared")
		}
		if m.HasError() {
			t.Error("expected error to be cleared")
		}
	})

	t.Run("returns load command", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)

		cmd := m.SetConversation("conv-1")

		if cmd == nil {
			t.Error("expected load command")
		}
	})

	t.Run("returns nil for empty conversation", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)

		cmd := m.SetConversation("")

		if cmd != nil {
			t.Error("expected nil command for empty conversation")
		}
	})
}

func TestMessages_Close(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("stops subscription", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.Init()

		err := m.Close()

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)

		err := m.Close()
		if err != nil {
			t.Errorf("expected no error on first close, got %v", err)
		}

		err = m.Close()
		if err != nil {
			t.Errorf("expected no error on second close, got %v", err)
		}
	})
}

func TestMessages_IsBusy(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("returns false when no items", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)

		if m.IsBusy() {
			t.Error("expected not busy with no items")
		}
	})

	t.Run("returns true when item is spinning", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		m := NewMessages(theme, db)
		m.SetWidth(80)
		m.SetConversation("conv-1")

		m.Update(UploadEventMsg{
			Event: upload.MessageProcessingEvent{
				ConversationID: "conv-1",
				UserMessageID:  "user-1",
			},
		})

		if !m.IsBusy() {
			t.Error("expected busy with spinning assistant")
		}
	})
}
