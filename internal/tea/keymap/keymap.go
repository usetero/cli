// Package keymap provides global key bindings and helpers for building keymaps.
package keymap

import "charm.land/bubbles/v2/key"

// Common key bindings.
var (
	Quit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)
	Exit = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("", ""),
	)
	Tab = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch focus"),
	)
	Send = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "send"),
	)
	Newline = key.NewBinding(
		key.WithKeys("shift+enter", "alt+enter", "ctrl+j"),
		key.WithHelp("shift+enter", "newline"),
	)
	Details = key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "details"),
	)
	// Global is shown in the keybar across all screens.
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
