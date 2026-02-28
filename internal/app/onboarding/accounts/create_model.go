package accounts

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/input"
)

// CreateModel handles account creation.
type CreateModel struct {
	ctx      context.Context
	theme    styles.Theme
	services api.APIServices
	prefs    preferences.OrgPreferences
	scope    log.Scope
	org      domain.Organization

	input    *input.Model
	creating bool
	err      error
	width    int
	height   int
}

// NewCreate creates a new account creation step.
func NewCreate(
	ctx context.Context,
	theme styles.Theme,
	org domain.Organization,
	services api.APIServices,
	prefs preferences.OrgPreferences,
	scope log.Scope,
) *CreateModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}

	inp := input.New(theme)
	inp.SetPlaceholder("Account name (e.g., Production, Staging)")
	inp.SetCharLimit(100)

	return &CreateModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		scope:    scope,
		org:      org,
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
	bindings := m.input.ShortHelp()
	bindings = append(bindings,
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "create")),
	)
	return bindings
}

func (m *CreateModel) createAccount(name string) tea.Cmd {
	return func() tea.Msg {
		id := uuid.New()
		account, err := m.services.Accounts.Create(m.ctx, api.CreateAccountInput{
			ID:             id,
			OrganizationID: m.org.ID,
			Name:           name,
		})
		if err != nil {
			return accountCreatedMsg{err: err}
		}
		return accountCreatedMsg{account: *account}
	}
}
