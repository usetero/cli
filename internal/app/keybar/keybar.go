// Package keybar renders keybinding hints at the bottom of the app.
package keybar

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/keymap"
)

// Model renders keybinding hints.
type Model struct {
	theme  styles.Theme
	scope  log.Scope
	help   help.Model
	keyMap help.KeyMap
	width  int
}

// New creates a new keybar.
func New(theme styles.Theme, scope log.Scope) *Model {
	scope = scope.Child("keybar")
	h := help.New()
	h.Styles = help.Styles{
		ShortKey:       lipgloss.NewStyle().Foreground(theme.TextMuted).Background(theme.Bg),
		ShortDesc:      lipgloss.NewStyle().Foreground(theme.TextSubtle).Background(theme.Bg),
		ShortSeparator: lipgloss.NewStyle().Foreground(theme.Border).Background(theme.Bg),
		Ellipsis:       lipgloss.NewStyle().Foreground(theme.Border).Background(theme.Bg),
		FullKey:        lipgloss.NewStyle().Foreground(theme.TextMuted).Background(theme.Bg),
		FullDesc:       lipgloss.NewStyle().Foreground(theme.TextSubtle).Background(theme.Bg),
		FullSeparator:  lipgloss.NewStyle().Foreground(theme.Border).Background(theme.Bg),
	}

	return &Model{
		theme: theme,
		scope: scope,
		help:  h,
	}
}

// SetWidth sets the keybar width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.help.SetWidth(width)
}

// SetKeyBindings sets the key bindings to display.
func (m *Model) SetKeyBindings(bindings []key.Binding) {
	if len(bindings) == 0 {
		m.keyMap = nil
		return
	}
	m.keyMap = keymap.Simple{Keys: bindings}
}

// View renders the keybar.
func (m *Model) View() string {
	if m.keyMap == nil {
		return ""
	}
	return m.help.ShortHelpView(m.keyMap.ShortHelp())
}

// Height returns the rendered height of the keybar.
func (m *Model) Height() int {
	if m.keyMap == nil {
		return 0
	}
	return 1
}
