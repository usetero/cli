// Package role provides the role selection step for onboarding.
package role

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
)

// Model handles role selection.
type Model struct {
	theme    styles.Theme
	prefs    preferences.UserPreferences
	scope    log.Scope
	selected int
	width    int
	height   int
}

// New creates a new role selection step.
func New(theme styles.Theme, prefs preferences.UserPreferences, scope log.Scope) *Model {
	if prefs == nil {
		panic("prefs is nil")
	}

	return &Model{
		theme: theme,
		prefs: prefs,
		scope: scope,
	}
}

// Init checks for a saved role preference.
func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		return savedRoleLoadedMsg{role: m.prefs.GetRole()}
	}
}

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "select")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	}
}
