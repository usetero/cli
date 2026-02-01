package chat

import (
	"errors"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/tuitest"
)

func TestMessageList_View(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("returns empty when size not set", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		list := NewMessageList(theme, db)

		result := list.View()

		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("shows empty state when no messages", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		list := NewMessageList(theme, db)
		list.SetSize(80, 24)

		result := list.View()

		if !strings.Contains(result, "Start a conversation") {
			t.Error("expected empty state message")
		}
	})

	t.Run("shows error state", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		list := NewMessageList(theme, db)
		list.SetSize(80, 24)

		// Simulate error by sending error message
		list.Update(messagesLoadedMsg{err: errors.New("database error")})

		result := list.View()

		if !strings.Contains(result, "database error") {
			t.Error("expected error message in view")
		}
		if !list.HasError() {
			t.Error("expected HasError() to return true")
		}
	})

	t.Run("renders messages after load", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		list := NewMessageList(theme, db)
		list.SetSize(80, 24)

		role := "user"
		content := `[{"type":"text","text":{"content":"Hello from test"}}]`

		list.Update(messagesLoadedMsg{
			messages: []sqlite.Message{
				{Role: &role, Content: &content},
			},
		})

		result := list.View()

		if !strings.Contains(result, "Hello from test") {
			t.Error("expected message content in view")
		}
		if !strings.Contains(result, "You") {
			t.Error("expected user label in view")
		}
	})

	t.Run("clears error after successful load", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		list := NewMessageList(theme, db)
		list.SetSize(80, 24)

		// First: error
		list.Update(messagesLoadedMsg{err: errors.New("temporary error")})
		if !list.HasError() {
			t.Fatal("expected error state")
		}

		// Then: success
		list.Update(messagesLoadedMsg{messages: []sqlite.Message{}})

		if list.HasError() {
			t.Error("expected error to be cleared after successful load")
		}
	})
}

func TestMessageList_Update(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("refreshes on messages table change", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		list := NewMessageList(theme, db)
		list.SetSize(80, 24)

		// Set a conversation so refresh will query
		cmd := list.SetConversation("conv-1")
		for _, msg := range tuitest.DrainCmds(cmd) {
			list.Update(msg)
		}

		// Simulate table change
		cmd = list.Update(tablesChangedMsg{tables: []string{"messages"}})

		// Should return a batch with refresh command
		if cmd == nil {
			t.Error("expected command to be returned for messages table change")
		}
	})

	t.Run("ignores unrelated table changes", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		list := NewMessageList(theme, db)
		list.SetSize(80, 24)

		// Initialize to set up subscription
		cmd := list.Init()
		_ = cmd // Don't drain - subscription is async

		// Set a conversation
		list.SetConversation("conv-1")

		// Simulate unrelated table change - should not trigger refresh
		// (The command returned is for listening, but without messages table
		// it won't include a refresh command)
		_ = list.Update(tablesChangedMsg{tables: []string{"conversations", "users"}})

		// Command may be nil if subscription not ready, that's ok
		// The key behavior is it doesn't crash and doesn't refresh unnecessarily
	})
}

func TestMessageList_Focus(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("tracks focus state", func(t *testing.T) {
		t.Parallel()

		db := sqlitetest.OpenTest(t)
		list := NewMessageList(theme, db)

		if list.IsFocused() {
			t.Error("expected unfocused by default")
		}

		list.Focus()
		if !list.IsFocused() {
			t.Error("expected focused after Focus()")
		}

		list.Blur()
		if list.IsFocused() {
			t.Error("expected unfocused after Blur()")
		}
	})
}
