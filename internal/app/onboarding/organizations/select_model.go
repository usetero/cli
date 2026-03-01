// Package organizations provides organization selection and creation steps.
package organizations

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/auth"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/loader"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

// SelectModel handles organization selection.
type SelectModel struct {
	ctx      context.Context
	theme    styles.Theme
	services graphql.ServiceSet
	prefs    preferences.UserPreferences
	auth     auth.Auth
	scope    log.Scope

	list            *remotelist.Model
	refreshLoader   *loader.Model
	orgs            []domain.Organization
	selectedOrg     *domain.Organization
	refreshingToken bool
	width           int
	height          int
}

// NewSelect creates a new organization select step.
func NewSelect(
	ctx context.Context,
	theme styles.Theme,
	services graphql.ServiceSet,
	prefs preferences.UserPreferences,
	authService auth.Auth,
	scope log.Scope,
) *SelectModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}
	if authService == nil {
		panic("authService is nil")
	}

	return &SelectModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		auth:     authService,
		scope:    scope,
		list:     remotelist.New(theme, "Loading organizations"),
	}
}

// Init starts loading organizations.
func (m *SelectModel) Init() tea.Cmd {
	m.scope.Info("loading organizations")
	return m.list.InitWithLoader(m.loadOrgs())
}

// SetSize updates dimensions.
func (m *SelectModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetWidth(width)
}

// ShortHelp returns the key bindings for the short help view.
func (m *SelectModel) ShortHelp() []key.Binding {
	if m.refreshingToken || m.list.IsLoading() {
		return nil
	}
	if m.list.HasError() {
		return []key.Binding{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		}
	}
	// Delegate to list, add step-specific bindings.
	bindings := m.list.ShortHelp()
	bindings = append(bindings,
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new org")),
	)
	return bindings
}
