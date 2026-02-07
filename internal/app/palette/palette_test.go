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
}
