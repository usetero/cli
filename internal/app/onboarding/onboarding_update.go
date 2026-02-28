package onboarding

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Update handles messages and orchestrates step transitions.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.step != nil {
			m.step.SetSize(m.width, m.height)
		}
		return nil
	}

	if cmd := m.handleTransition(msg); cmd != nil {
		return cmd
	}

	// Delegate to current step
	if m.step != nil {
		return m.step.Update(msg)
	}
	return nil
}

// SetSize updates the model's dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	if m.step != nil {
		m.step.SetSize(width, height)
	}
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	if m.step != nil {
		if h, ok := m.step.(HelpProvider); ok {
			return h.ShortHelp()
		}
	}
	return nil
}
