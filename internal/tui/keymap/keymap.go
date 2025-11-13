package keymap

import "github.com/charmbracelet/bubbles/v2/key"

// Simple implements help.KeyMap with a basic list of key bindings.
// Use this when you need to return dynamic keymaps from Help() methods
// in steps, pages, or layouts.
type Simple struct {
	Keys []key.Binding
}

// ShortHelp returns the key bindings for short help view
func (k Simple) ShortHelp() []key.Binding {
	return k.Keys
}

// FullHelp returns the key bindings for full help view
func (k Simple) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.Keys}
}
