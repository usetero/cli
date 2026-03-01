package organizations

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	graphql "github.com/usetero/cli/internal/boundary/graphql"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/input"
)

// CreateModel handles organization creation.
type CreateModel struct {
	ctx      context.Context
	theme    styles.Theme
	services graphql.ServiceSet
	prefs    preferences.UserPreferences
	scope    log.Scope

	input    *input.Model
	creating bool
	err      error
	width    int
	height   int
}

// NewCreate creates a new organization creation step.
func NewCreate(
	ctx context.Context,
	theme styles.Theme,
	services graphql.ServiceSet,
	prefs preferences.UserPreferences,
	scope log.Scope,
) *CreateModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}

	inp := input.New(theme)
	inp.SetPlaceholder("Organization name")
	inp.SetCharLimit(100)

	return &CreateModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		scope:    scope,
		input:    inp,
	}
}

// Init focuses the input.
func (m *CreateModel) Init() tea.Cmd {
	return m.input.Focus()
}

// SetSize updates dimensions.
func (m *CreateModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.SetWidth(width)
}

// ShortHelp returns the key bindings for the short help view.
func (m *CreateModel) ShortHelp() []key.Binding {
	if m.creating {
		return nil
	}
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "create")),
	}
}
