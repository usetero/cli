// Package datadog provides Datadog configuration steps.
package datadog

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/api"
	appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// CheckModel checks if Datadog is configured for the account.
type CheckModel struct {
	ctx      context.Context
	theme    styles.Theme
	services api.APIServices
	scope    log.Scope
	account  domain.Account
	err      error
}

// NewCheck creates a new datadog check step.
func NewCheck(
	ctx context.Context,
	theme styles.Theme,
	account domain.Account,
	services api.APIServices,
	scope log.Scope,
) *CheckModel {
	if ctx == nil {
		panic("ctx is nil")
	}

	return &CheckModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		scope:    scope,
		account:  account,
	}
}

// Init starts checking for Datadog configuration.
func (m *CheckModel) Init() tea.Cmd {
	m.scope.Info("checking datadog configuration", "accountID", m.account.ID)
	return m.checkDatadog()
}

// SetSize updates dimensions.
func (m *CheckModel) SetSize(width, height int) {}

// ShortHelp returns the key bindings for the short help view.
func (m *CheckModel) ShortHelp() []key.Binding {
	if m.err != nil {
		return []key.Binding{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		}
	}
	return nil
}

func (m *CheckModel) Hidden() bool {
	return m.err == nil
}

func (m *CheckModel) Status() appmsg.StepStatus {
	return appmsg.StepStatus{
		Title:   "Datadog setup",
		Details: "Checking Datadog configuration...",
	}
}
