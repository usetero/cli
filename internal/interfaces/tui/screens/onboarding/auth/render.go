package auth

import tea "charm.land/bubbletea/v2"

// SetSize satisfies the shared TUI model contract.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// View renders the auth onboarding body.
func (m *Model) View() tea.View {
	return tea.NewView(" ")
}
