// Package accounts provides account selection and creation steps.
package accounts

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

// SelectModel handles account selection.
type SelectModel struct {
	ctx      context.Context
	theme    *styles.Theme
	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger
	org      domain.Organization

	list     *remotelist.Model
	accounts []domain.Account
	width    int
	height   int
}

// NewSelect creates a new account select step.
func NewSelect(
	ctx context.Context,
	theme *styles.Theme,
	org domain.Organization,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) *SelectModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}
	return &SelectModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		logger:   logger,
		org:      org,
		list:     remotelist.New(theme, "Loading accounts"),
	}
}

// Init starts loading accounts.
func (m *SelectModel) Init() tea.Cmd {
	m.logger.Info("loading accounts", "orgID", m.org.ID)
	return m.list.InitWithLoader(m.loadAccounts())
}

func (m *SelectModel) loadAccounts() tea.Cmd {
	return func() tea.Msg {
		accounts, err := m.services.Accounts.List(m.ctx, m.org.ID)
		if err != nil {
			return remotelist.LoadResult{Err: err}
		}

		items := make([]remotelist.Item, len(accounts))
		for i, acc := range accounts {
			items[i] = acc
		}
		return remotelist.LoadResult{Items: items}
	}
}

// Update handles messages.
func (m *SelectModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case remotelist.LoadResult:
		if msg.Err != nil {
			m.logger.Error("failed to load accounts", "error", msg.Err)
			return m.list.Update(msg)
		}

		m.accounts = make([]domain.Account, len(msg.Items))
		for i, item := range msg.Items {
			m.accounts[i] = item.(domain.Account)
		}
		m.logger.Info("accounts loaded", "count", len(m.accounts))

		if len(m.accounts) == 0 {
			m.logger.Debug("no accounts found")
			org := m.org
			return func() tea.Msg { return msgs.NoAccounts{Org: org} }
		}

		prefID := m.prefs.GetDefaultAccountID()
		if prefID != "" {
			for _, acc := range m.accounts {
				if acc.ID == prefID {
					m.logger.Debug("using saved account preference", "id", acc.ID)
					m.services.SetAccountID(acc.ID)
					return m.emitSelected(acc)
				}
			}
		}

		if len(m.accounts) == 1 {
			m.logger.Debug("auto-selected account (only one)")
			acc := m.accounts[0]
			_ = m.prefs.SetDefaultAccountID(acc.ID)
			m.services.SetAccountID(acc.ID)
			return m.emitSelected(acc)
		}

		return m.list.Update(msg)

	case tea.KeyPressMsg:
		if m.list.IsLoading() {
			return nil
		}
		switch msg.String() {
		case "enter":
			if item := m.list.SelectedItem(); item != nil {
				acc := item.(domain.Account)
				_ = m.prefs.SetDefaultAccountID(acc.ID)
				m.services.SetAccountID(acc.ID)
				return m.emitSelected(acc)
			}
		case "n":
			org := m.org
			return func() tea.Msg { return msgs.NoAccounts{Org: org} }
		case "r":
			if m.list.HasError() {
				m.logger.Debug("retrying account load")
				return m.list.Retry()
			}
		}
	}

	return m.list.Update(msg)
}

func (m *SelectModel) emitSelected(acc domain.Account) tea.Cmd {
	org := m.org
	m.logger.Info("account selected", "id", acc.ID, "name", acc.Name)
	return func() tea.Msg {
		return msgs.AccountSelected{Org: org, Account: acc}
	}
}

// View renders the account selection UI.
func (m *SelectModel) View() string {
	s := m.theme.Styles

	if m.list.IsLoading() {
		return m.list.View()
	}

	if m.list.HasError() {
		return s.Error.Render("Failed to load accounts. Press 'r' to retry.")
	}

	title := s.Title.Render("Select your Datadog account")
	subtitle := s.Help.Render("Press 'n' to connect a new Datadog account")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		m.list.View(),
	)
}

// SetSize updates dimensions.
func (m *SelectModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetWidth(width)
}
