package palette

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/teatest"
)

func testCommands() []Command {
	noop := func() tea.Cmd { return nil }
	return []Command{
		{Name: "New Conversation", Handler: noop},
		{Name: "Toggle Details", Handler: noop},
		{Name: "Theme", Children: []Command{
			{Name: "Auto", Handler: noop},
			{Name: "Dark", Handler: noop},
			{Name: "Light", Handler: noop},
		}},
		{Name: "Quit", Handler: noop},
	}
}

func TestView(t *testing.T) {
	t.Parallel()
	theme := styles.NewTheme(true)

	t.Run("no raw escape sequences", func(t *testing.T) {
		t.Parallel()
		m := New(theme, testCommands())
		m.SetWidth(50)

		output := m.View()
		t.Logf("output:\n%s", output)
		teatest.AssertNoRawEscapes(t, output)
	})

	t.Run("respects width", func(t *testing.T) {
		t.Parallel()
		for _, width := range []int{30, 50, 80} {
			t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
				t.Parallel()
				m := New(theme, testCommands())
				m.SetWidth(width)

				teatest.AssertMaxWidth(t, width, m.View())
			})
		}
	})

	t.Run("no raw escapes after filtering", func(t *testing.T) {
		t.Parallel()
		m := New(theme, testCommands())
		m.SetWidth(50)

		m.input.SetValue("qu")
		m.filter()

		if len(m.matches) != 1 {
			t.Fatalf("expected 1 match, got %d", len(m.matches))
		}

		output := m.View()
		teatest.AssertNoRawEscapes(t, output)
	})

	t.Run("drill into children shows child commands", func(t *testing.T) {
		t.Parallel()
		m := New(theme, testCommands())
		m.SetWidth(50)

		// Select "Theme" (index 2) and press enter to drill in
		m.selected = 2
		m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Should now show 3 child commands
		if len(m.matches) != 3 {
			t.Fatalf("expected 3 children, got %d", len(m.matches))
		}
		if m.matches[0].command.Name != "Auto" {
			t.Fatalf("expected first child 'Auto', got %q", m.matches[0].command.Name)
		}

		output := m.View()
		teatest.AssertNoRawEscapes(t, output)
		teatest.AssertMaxWidth(t, 50, output)
	})

	t.Run("escape from children returns to parent", func(t *testing.T) {
		t.Parallel()
		m := New(theme, testCommands())
		m.SetWidth(50)

		// Drill into Theme
		m.selected = 2
		m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		if len(m.stack) != 1 {
			t.Fatalf("expected stack depth 1, got %d", len(m.stack))
		}

		// Escape should go back, not close
		cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		if cmd != nil {
			t.Fatal("expected nil cmd (back), got non-nil (would close)")
		}
		if len(m.stack) != 0 {
			t.Fatalf("expected stack depth 0, got %d", len(m.stack))
		}
		if len(m.matches) != len(testCommands()) {
			t.Fatalf("expected %d top-level matches, got %d", len(testCommands()), len(m.matches))
		}
	})
}
