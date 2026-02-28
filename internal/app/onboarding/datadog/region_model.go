package datadog

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// RegionModel handles Datadog region selection.
type RegionModel struct {
	theme    styles.Theme
	scope    log.Scope
	selected int
	width    int
	height   int
}

// NewRegion creates a new region selection step.
func NewRegion(theme styles.Theme, scope log.Scope) *RegionModel {

	return &RegionModel{
		theme: theme,
		scope: scope,
	}
}

// Init returns nil.
func (m *RegionModel) Init() tea.Cmd {
	return nil
}

// SetSize updates dimensions.
func (m *RegionModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ShortHelp returns the key bindings for the short help view.
func (m *RegionModel) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "select")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	}
}
