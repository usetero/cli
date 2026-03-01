// Package runtimeinit ensures account runtime dependencies are initialized.
package runtimeinit

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	onbstatus "github.com/usetero/cli/internal/app/onboarding/status"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

type Model struct {
	theme   styles.Theme
	scope   log.Scope
	org     domain.Organization
	account domain.Account
}

func New(theme styles.Theme, org domain.Organization, account domain.Account, scope log.Scope) *Model {
	return &Model{
		theme:   theme,
		scope:   scope,
		org:     org,
		account: account,
	}
}

func (m *Model) Init() tea.Cmd {
	m.scope.Debug("requesting runtime initialization", "account_id", m.account.ID.String())
	return func() tea.Msg {
		return bootstrap.EnsureRuntime{Org: m.org, Account: m.account}
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd { return nil }

func (m *Model) View() string {
	s := m.theme.Styles
	return lipgloss.JoinVertical(
		lipgloss.Left,
		s.Title.Render("Getting ready"),
		"",
		s.Body.Render("Initializing your account runtime..."),
	)
}

func (m *Model) SetSize(width, height int) {}

func (m *Model) ShortHelp() []key.Binding { return nil }

func (m *Model) Hidden() bool { return true }

func (m *Model) Status() onbstatus.StepStatus {
	return onbstatus.StepStatus{
		Title:   "Getting ready",
		Details: "Initializing your account runtime...",
	}
}
