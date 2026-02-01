package chat_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat"
	"github.com/usetero/cli/internal/tui/app/chat/chattest"
)

func TestMessageList_View(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("returns empty when size not set", func(t *testing.T) {
		t.Parallel()

		mock := &chattest.MockMessages{}
		list := chat.NewMessageList(theme, mock)

		result := list.View()

		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("shows empty state when no items", func(t *testing.T) {
		t.Parallel()

		mock := &chattest.MockMessages{
			ItemsFunc: func() []chat.Item { return nil },
		}
		list := chat.NewMessageList(theme, mock)
		list.SetSize(80, 24)

		result := list.View()

		if !strings.Contains(result, "Start a conversation") {
			t.Error("expected empty state message")
		}
	})

	t.Run("shows error state", func(t *testing.T) {
		t.Parallel()

		mock := &chattest.MockMessages{
			HasErrorFunc: func() bool { return true },
			ErrorFunc:    func() error { return errors.New("database error") },
		}
		list := chat.NewMessageList(theme, mock)
		list.SetSize(80, 24)

		result := list.View()

		if !strings.Contains(result, "database error") {
			t.Error("expected error message in view")
		}
	})

	t.Run("renders items from messages", func(t *testing.T) {
		t.Parallel()

		userMsg := chat.NewUserMessage(theme, "msg-1")
		userMsg.SetWidth(80)

		mock := &chattest.MockMessages{
			ItemsFunc: func() []chat.Item { return []chat.Item{userMsg} },
		}
		list := chat.NewMessageList(theme, mock)
		list.SetSize(80, 24)

		result := list.View()

		if !strings.Contains(result, "You") {
			t.Error("expected user label in view")
		}
	})
}

func TestMessageList_Update(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("delegates to messages", func(t *testing.T) {
		t.Parallel()

		called := false
		mock := &chattest.MockMessages{
			UpdateFunc: func(msg tea.Msg) tea.Cmd {
				called = true
				return nil
			},
		}
		list := chat.NewMessageList(theme, mock)

		list.Update("test message")

		if !called {
			t.Error("expected Update to be delegated to messages")
		}
	})

	t.Run("returns command from messages", func(t *testing.T) {
		t.Parallel()

		expectedCmd := func() tea.Msg { return "cmd" }
		mock := &chattest.MockMessages{
			UpdateFunc: func(msg tea.Msg) tea.Cmd { return expectedCmd },
		}
		list := chat.NewMessageList(theme, mock)

		cmd := list.Update("test")

		if cmd == nil {
			t.Error("expected command to be returned")
		}
	})
}

func TestMessageList_Focus(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("tracks focus state", func(t *testing.T) {
		t.Parallel()

		mock := &chattest.MockMessages{}
		list := chat.NewMessageList(theme, mock)

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

	t.Run("shows focus indicator in view", func(t *testing.T) {
		t.Parallel()

		userMsg := chat.NewUserMessage(theme, "msg-1")
		mock := &chattest.MockMessages{
			ItemsFunc: func() []chat.Item { return []chat.Item{userMsg} },
		}
		list := chat.NewMessageList(theme, mock)
		list.SetSize(80, 24)

		list.Focus()
		focused := list.View()

		list.Blur()
		blurred := list.View()

		// Focused view should be different (has border)
		if focused == blurred {
			t.Error("expected focus indicator to change view")
		}
	})
}

func TestMessageList_SetSize(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("sets width on messages", func(t *testing.T) {
		t.Parallel()

		var widthSet int
		mock := &chattest.MockMessages{
			SetWidthFunc: func(width int) { widthSet = width },
		}
		list := chat.NewMessageList(theme, mock)

		list.SetSize(100, 50)

		// Width passed to messages should be width - 2 (for padding)
		if widthSet != 98 {
			t.Errorf("expected width 98, got %d", widthSet)
		}
	})
}

func TestMessageList_SetConversation(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("delegates to messages", func(t *testing.T) {
		t.Parallel()

		var convID string
		mock := &chattest.MockMessages{
			SetConversationFunc: func(id string) tea.Cmd {
				convID = id
				return nil
			},
		}
		list := chat.NewMessageList(theme, mock)

		list.SetConversation("conv-123")

		if convID != "conv-123" {
			t.Errorf("expected conversation ID 'conv-123', got %q", convID)
		}
	})
}

func TestMessageList_Delegates(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("IsBusy delegates to messages", func(t *testing.T) {
		t.Parallel()

		mock := &chattest.MockMessages{
			IsBusyFunc: func() bool { return true },
		}
		list := chat.NewMessageList(theme, mock)

		if !list.IsBusy() {
			t.Error("expected IsBusy to delegate to messages")
		}
	})

	t.Run("HasError delegates to messages", func(t *testing.T) {
		t.Parallel()

		mock := &chattest.MockMessages{
			HasErrorFunc: func() bool { return true },
		}
		list := chat.NewMessageList(theme, mock)

		if !list.HasError() {
			t.Error("expected HasError to delegate to messages")
		}
	})

	t.Run("Close delegates to messages", func(t *testing.T) {
		t.Parallel()

		closed := false
		mock := &chattest.MockMessages{
			CloseFunc: func() error {
				closed = true
				return nil
			},
		}
		list := chat.NewMessageList(theme, mock)

		list.Close()

		if !closed {
			t.Error("expected Close to delegate to messages")
		}
	})

	t.Run("Refresh delegates to messages", func(t *testing.T) {
		t.Parallel()

		refreshed := false
		mock := &chattest.MockMessages{
			RefreshFunc: func() tea.Cmd {
				refreshed = true
				return nil
			},
		}
		list := chat.NewMessageList(theme, mock)

		list.Refresh()

		if !refreshed {
			t.Error("expected Refresh to delegate to messages")
		}
	})

	t.Run("Init delegates to messages", func(t *testing.T) {
		t.Parallel()

		initialized := false
		mock := &chattest.MockMessages{
			InitFunc: func() tea.Cmd {
				initialized = true
				return nil
			},
		}
		list := chat.NewMessageList(theme, mock)

		list.Init()

		if !initialized {
			t.Error("expected Init to delegate to messages")
		}
	})
}
