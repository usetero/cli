// Package workspaces provides workspace selection steps.
package workspaces

import (
	"context"
	"log/slog"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

// SelectModel handles workspace selection.
type SelectModel struct {
	ctx      context.Context
	theme    styles.Theme
	services graphql.ServiceSet
	prefs    preferences.OrgPreferences
	account  domain.Account
	scope    log.Scope

	list       *remotelist.Model
	workspaces []domain.Workspace
	width      int
	height     int
}

// NewSelect creates a new workspace select step.
func NewSelect(
	ctx context.Context,
	theme styles.Theme,
	account domain.Account,
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

	scope.Debug("initialized")

	return &SelectModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		account:  account,
		scope:    scope,
		list:     remotelist.New(theme, "Loading workspaces"),
	}
}

// Init starts loading workspaces.
func (m *SelectModel) Init() tea.Cmd {
	m.scope.Debug("loading workspaces", slog.String("account_id", m.account.ID.String()))
	return m.list.InitWithLoader(m.loadWorkspaces())
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
	)
	return bindings
}

func (m *SelectModel) Hidden() bool {
	return m.list.IsLoading()
}

func (m *SelectModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{
		Title:   "Select workspace",
		Details: "Loading workspaces...",
	}
}
