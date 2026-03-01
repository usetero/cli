package inputbar

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	msgs "github.com/usetero/cli/internal/app/chat/events"
	"github.com/usetero/cli/internal/app/palette"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func newTestInputBar(t *testing.T) *Model {
	t.Helper()
	m := New(nil, styles.NewTheme(true), logtest.NewScope(t))
	m.SetWidth(80)
	return m
}

func TestUpdate_Submit(t *testing.T) {
	t.Parallel()
	m := newTestInputBar(t)
	m.textarea.SetValue(" hello ")

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected submit cmd")
	}
	msg := cmd()
	input, ok := msg.(msgs.UserSubmittedInput)
	if !ok {
		t.Fatalf("expected UserSubmittedInput, got %T", msg)
	}
	if input.Text != "hello" {
		t.Fatalf("submitted text = %q, want hello", input.Text)
	}
	if m.textarea.Value() != "" {
		t.Fatalf("textarea should reset, got %q", m.textarea.Value())
	}
}

func TestUpdate_Newline(t *testing.T) {
	t.Parallel()
	m := newTestInputBar(t)
	m.textarea.SetValue("a")

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModShift})
	if cmd != nil {
		t.Fatal("newline should not emit cmd")
	}
	if !strings.Contains(m.textarea.Value(), "\n") {
		t.Fatalf("expected newline in textarea, got %q", m.textarea.Value())
	}
}

func TestUpdate_Palette(t *testing.T) {
	t.Parallel()
	m := newTestInputBar(t)

	cmd := m.Update(tea.KeyPressMsg{Text: "/"})
	if cmd == nil {
		t.Fatal("expected palette open cmd")
	}
	if _, ok := cmd().(palette.OpenMsg); !ok {
		t.Fatalf("expected palette.OpenMsg, got %T", cmd())
	}
}

func TestUpdate_PendingTextRestore(t *testing.T) {
	t.Parallel()
	m := newTestInputBar(t)
	m.pendingText = "restored text"

	m.Update(msgs.StreamFailed{Err: nil})
	if m.textarea.Value() != "restored text" {
		t.Fatalf("textarea = %q, want restored text", m.textarea.Value())
	}

	m.Update(msgs.StreamCompleted{})
	if m.pendingText != "" {
		t.Fatalf("pendingText = %q, want empty", m.pendingText)
	}
}
