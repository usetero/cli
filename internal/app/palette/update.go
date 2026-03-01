package palette

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
)

// Update handles key events. The caller must only forward events when the palette is open.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, closeKey):
			// If nested, go back one level instead of closing.
			if len(m.stack) > 0 {
				m.popLevel()
				return nil
			}
			return func() tea.Msg { return appevents.PaletteCloseRequested{} }

		case key.Matches(msg, selectKey):
			if len(m.matches) > 0 {
				cmd := m.matches[m.selected].command
				// If command has children, drill down.
				if len(cmd.Children) > 0 {
					m.pushLevel(cmd.Name, cmd.Children)
					return nil
				}
				return tea.Sequence(
					func() tea.Msg { return appevents.PaletteCloseRequested{} },
					cmd.Handler(),
				)
			}
			return nil

		case key.Matches(msg, nextKey):
			if m.selected < len(m.matches)-1 {
				m.selected++
			}
			return nil

		case key.Matches(msg, prevKey):
			if m.selected > 0 {
				m.selected--
			}
			return nil
		}
	}

	// Forward to text input for typing.
	prevValue := m.input.Value()
	cmd := m.input.Update(msg)
	if m.input.Value() != prevValue {
		m.filter()
	}
	return cmd
}
