// Package accounts provides account selection and creation steps.
package accounts

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

// SelectModel handles account selection.
type SelectModel struct {
	ctx      context.Context
	theme    styles.Theme
	services graphql.ServiceSet
	prefs    preferences.OrgPreferences
	scope    log.Scope
	org      domain.Organization

	list     *remotelist.Model
	accounts []domain.Account
	width    int
	height   int
}

// NewSelect creates a new account select step.
func NewSelect(
	ctx context.Context,
	theme styles.Theme,
	org domain.Organization,
	services graphql.ServiceSet,
	prefs preferences.OrgPreferences,
	scope log.Scope,
) *SelectModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}

	return &SelectModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		scope:    scope,
		org:      org,
		list:     remotelist.New(theme, "Loading accounts"),
	}
}

// Init starts loading accounts.
func (m *SelectModel) Init() tea.Cmd {
	m.scope.Info("loading accounts", "orgID", m.org.ID)
	return m.list.InitWithLoader(m.loadAccounts())
}

// SetSize updates dimensions.
func (m *SelectModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetWidth(width)
}

// ShortHelp returns the key bindings for the short help view.
func (m *SelectModel) ShortHelp() []key.Binding {
	if m.list.IsLoading() {
		return nil
	}
	if m.list.HasError() {
		return []key.Binding{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		}
	}
	bindings := m.list.ShortHelp()
	bindings = append(bindings,
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new account")),
	)
	return bindings
}
