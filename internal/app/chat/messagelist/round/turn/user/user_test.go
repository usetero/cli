package user

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	msgs "github.com/usetero/cli/internal/app/chat/events"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/styles"
)

func newTestUser(t *testing.T, text string, width int) *Model {
	t.Helper()
	theme := styles.NewTheme(true)
	return New(theme, "user-1", msgs.UserSubmittedInput{Text: text}, width)
}

func TestView(t *testing.T) {
	t.Parallel()

	t.Run("renders within width", func(t *testing.T) {
		t.Parallel()
		for _, width := range []int{40, 60, 80, 120} {
			m := newTestUser(t, "Hello, world!", width)
			view := m.View()
			for i, line := range strings.Split(view, "\n") {
				w := ansi.StringWidth(line)
				if w > width {
					t.Errorf("width=%d line %d: got %d chars, want <=%d", width, i, w, width)
				}
			}
		}
	})

	t.Run("wraps long text", func(t *testing.T) {
		t.Parallel()
		long := strings.Repeat("word ", 30) // ~150 chars
		m := newTestUser(t, long, 40)
		h := m.Height()
		if h < 2 {
			t.Errorf("expected wrapping (height >= 2), got height %d", h)
		}
	})

	t.Run("tool results are invisible", func(t *testing.T) {
		t.Parallel()
		theme := styles.NewTheme(true)
		m := New(theme, "user-1", msgs.UserSubmittedInput{
			ToolResults: []tools.Result{{ToolUseID: "tool-1"}},
		}, 80)

		if m.View() != "" {
			t.Error("expected empty view for tool result message")
		}
		if m.Height() != 0 {
			t.Errorf("expected height 0 for tool result message, got %d", m.Height())
		}
		if m.IsVisible() {
			t.Error("expected IsVisible() == false for tool result message")
		}
	})
}

func TestKind(t *testing.T) {
	t.Parallel()
	m := newTestUser(t, "hi", 80)
	if m.Kind() != block.KindUser {
		t.Errorf("expected KindUser, got %d", m.Kind())
	}
}

func TestFocused(t *testing.T) {
	t.Parallel()
	m := newTestUser(t, "hi", 80)

	if m.Focused() {
		t.Error("expected unfocused by default")
	}
	m.SetFocused(true)
	if !m.Focused() {
		t.Error("expected focused after SetFocused(true)")
	}
}
