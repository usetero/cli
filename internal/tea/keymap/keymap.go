// Package keymap provides global key bindings and helpers for building keymaps.
package keymap

import "charm.land/bubbles/v2/key"

// Global key bindings available throughout the application.
var (
	Quit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)
	Exit = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("", ""), // Works but not shown in help
	)

	// Global is the set of global bindings shown in help.
	Global = []key.Binding{Quit}
)

// Simple implements help.KeyMap with a basic list of key bindings.
// Use this when you need to return dynamic keymaps from Help() methods.
type Simple struct {
	Keys []key.Binding
}

// ShortHelp returns the key bindings for the short help view.
func (k Simple) ShortHelp() []key.Binding {
	return k.Keys
}

// FullHelp returns the key bindings for the full help view.
func (k Simple) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.Keys}
}
