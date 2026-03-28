package commandbar

import "charm.land/bubbles/v2/key"

// ShortHelp exposes the active command-bar child bindings when that child owns interaction.
func (m *Model) ShortHelp() []key.Binding {
	if m.busy != nil {
		return nil
	}

	if child := m.active(); child != nil {
		type helpProvider interface {
			ShortHelp() []key.Binding
		}
		if provider, ok := child.(helpProvider); ok {
			return provider.ShortHelp()
		}
	}

	return nil
}
