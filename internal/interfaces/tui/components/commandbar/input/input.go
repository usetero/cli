package input

import "charm.land/bubbles/v2/key"

// ShortHelp satisfies the shared TUI model contract.
func (m *Model) ShortHelp() []key.Binding {
	if m.allowsNewline() {
		return []key.Binding{sendBinding, newlineBinding}
	}
	return []key.Binding{sendBinding}
}
