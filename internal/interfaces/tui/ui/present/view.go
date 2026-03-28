package present

import tea "charm.land/bubbletea/v2"

// View wraps already-rendered presentation content in a tea.View.
func View(content string) tea.View {
	return tea.NewView(content)
}
