// Package accounts provides account selection and creation steps.
package accounts

import (
	"context"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

const (
	accountsProvisioningMaxWait  = 60 * time.Second
	accountsProvisioningMaxTries = 6
)

type retryLoadAccountsMsg struct{}

// SelectModel handles account selection.
type SelectModel struct {
	ctx      context.Context
	theme    styles.Theme
	services api.APIServices
	prefs    preferences.OrgPreferences
	scope    log.Scope
	org      domain.Organization

	list     *remotelist.Model
	accounts []domain.Account
	width    int
	height   int

	// Retry state: handles short backend provisioning delays after org creation.
	accountsLoadStartedAt time.Time
	accountsLoadAttempts  int
}

// NewSelect creates a new account select step.
func NewSelect(
	ctx context.Context,
	theme styles.Theme,
	org domain.Organization,
	services api.APIServices,
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
		list:     remotelist.New(theme, "Loading accounts…"),
	}
}

// Init starts loading accounts.
func (m *SelectModel) Init() tea.Cmd {
	m.accountsLoadStartedAt = time.Now()
	m.accountsLoadAttempts = 0

	m.scope.Info("loading accounts", "orgID", m.org.ID)
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
	case retryLoadAccountsMsg:
		// Re-run the accounts loader.
		return m.list.InitWithLoader(m.loadAccounts())

	case remotelist.LoadResult:
		if msg.Err != nil {
			// Only auto-retry during a short window and only for the known provisioning symptom.
			errText := strings.ToLower(msg.Err.Error())
			if strings.Contains(errText, "accounts internal error") &&
				m.accountsLoadAttempts < accountsProvisioningMaxTries &&
				time.Since(m.accountsLoadStartedAt) < accountsProvisioningMaxWait {

				m.accountsLoadAttempts++
				delay := time.Duration(m.accountsLoadAttempts) * time.Second // 1s, 2s, 3s...

				m.scope.Info(
					"accounts not ready yet; retrying",
					"attempt", m.accountsLoadAttempts,
					"delay", delay.String(),
					"error", msg.Err,
				)

				// Keep the user in "loading" mode; don’t flip the list into an error state yet.
				return tea.Tick(delay, func(time.Time) tea.Msg { return retryLoadAccountsMsg{} })
			}

			m.scope.Error("failed to load accounts", "error", msg.Err)
			return tea.Batch(m.list.Update(msg), appmsg.ErrorCmd("Failed to load accounts", msg.Err, false))
		}

		m.accounts = make([]domain.Account, len(msg.Items))
		for i, item := range msg.Items {
			if acc, ok := item.(domain.Account); ok {
				m.accounts[i] = acc
			}
		}
		m.scope.Info("accounts loaded", "count", len(m.accounts))

		if len(m.accounts) == 0 {
			m.scope.Debug("no accounts found")
			org := m.org
			return func() tea.Msg { return msgs.NoAccounts{Org: org} }
		}

		prefID := m.prefs.GetDefaultAccountID()
		if prefID != "" {
			for _, acc := range m.accounts {
				if acc.ID == prefID {
					m.scope.Debug("using saved account preference", "id", acc.ID)
					m.services.SetAccountID(acc.ID)
					return m.emitSelected(acc)
				}
			}
		}

		if len(m.accounts) == 1 {
			m.scope.Debug("auto-selected account (only one)")
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
				if acc, ok := item.(domain.Account); ok {
					_ = m.prefs.SetDefaultAccountID(acc.ID)
					m.services.SetAccountID(acc.ID)
					return m.emitSelected(acc)
				}
			}
		case "n":
			org := m.org
			return func() tea.Msg { return msgs.NoAccounts{Org: org} }
		case "r":
			if m.list.HasError() {
				m.scope.Debug("retrying account load")
				return m.list.Retry()
			}
		}
	}

	return m.list.Update(msg)
}

func (m *SelectModel) emitSelected(acc domain.Account) tea.Cmd {
	org := m.org
	m.scope.Info("account selected", "id", acc.ID, "name", acc.Name)
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
