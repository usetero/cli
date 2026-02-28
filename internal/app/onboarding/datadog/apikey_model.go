package datadog

import (
	"context"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/input"
)

// APIKeyModel handles Datadog API key entry.
type APIKeyModel struct {
	ctx      context.Context
	theme    styles.Theme
	services api.APIServices
	scope    log.Scope
	account  domain.Account
	site     domain.DatadogSite

	input      *input.Model
	validating bool
	err        error
	width      int
	height     int
}

// NewAPIKey creates a new API key entry step.
func NewAPIKey(
	ctx context.Context,
	theme styles.Theme,
	account domain.Account,
	site domain.DatadogSite,
	services api.APIServices,
	scope log.Scope,
) *APIKeyModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if site == "" {
		panic("site is empty")
	}

	inp := input.New(theme)
	inp.SetPlaceholder("Datadog API Key")
	inp.SetCharLimit(64)
	inp.SetEchoMode(textinput.EchoPassword)
	inp.SetEchoCharacter('•')

	return &APIKeyModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		scope:    scope,
		account:  account,
		site:     site,
		input:    inp,
	}
}

// Init focuses the input.
func (m *APIKeyModel) Init() tea.Cmd {
	return m.input.Focus()
}

// SetSize updates dimensions.
func (m *APIKeyModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.SetWidth(width)
}

// ShortHelp returns the key bindings for the short help view.
func (m *APIKeyModel) ShortHelp() []key.Binding {
	if m.validating {
		return nil
	}
	bindings := m.input.ShortHelp()
	bindings = append(bindings,
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "validate")),
	)
	return bindings
}
