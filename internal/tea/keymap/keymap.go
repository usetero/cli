// Package keymap provides global key bindings and helpers for building keymaps.
package keymap

import "charm.land/bubbles/v2/key"

// Global bindings — always active.
var (
	Quit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)
	Exit = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("", ""),
	)

	// Global is shown in the keybar across all screens.
	Global = []key.Binding{Quit}
)

// App bindings — active during normal interaction.
var (
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
	Palette = key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "commands"),
	)
)

// Drawer bindings — active when the statusbar drawer is open.
var (
	NextTab = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	)
	CloseDrawer = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	)
	DrawerUp = key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	)
	DrawerDown = key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	)
	DrawerSelect = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	)
	DrawerBack = key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc/bksp", "back"),
	)
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
