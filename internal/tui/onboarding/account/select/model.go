package accountselect

import (
	"context"
	"fmt"
	"io"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	bubbleslist "charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/remotelist"
	"github.com/usetero/cli/internal/tui/keymap"
	accountcreate "github.com/usetero/cli/internal/tui/onboarding/account/create"
	datadogcheck "github.com/usetero/cli/internal/tui/onboarding/datadog/check"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

const createNewAccountID domain.AccountID = "__CREATE_NEW__"

// AccountSelectedMsg is sent when an account is selected during onboarding.
// This triggers the syncer to start in the background.
type AccountSelectedMsg struct {
	Organization domain.Organization
	Account      domain.Account
}

// AccountListItem wraps domain.Account to implement list.Item.
type AccountListItem struct {
	domain.Account
}

func (i AccountListItem) FilterValue() string { return i.Name }

// accountDelegate renders each account in the list.
type accountDelegate struct {
	theme *styles.Theme
}

func (d accountDelegate) Height() int                                    { return 1 }
func (d accountDelegate) Spacing() int                                   { return 0 }
func (d accountDelegate) Update(_ tea.Msg, _ *bubbleslist.Model) tea.Cmd { return nil }
func (d accountDelegate) Render(w io.Writer, m bubbleslist.Model, index int, item bubbleslist.Item) {
	i, ok := item.(AccountListItem)
	if !ok {
		return
	}

	colors := d.theme.Colors
	str := i.Name
	if index == m.Index() {
		fmt.Fprint(w, lipgloss.NewStyle().Foreground(colors.Accent).Bold(true).Render("> "+str))
	} else {
		fmt.Fprint(w, lipgloss.NewStyle().Foreground(colors.Page.Text).Render("  "+str))
	}
}

// Model handles selecting an account.
type Model struct {
	ctx   context.Context
	theme *styles.Theme
	role  string
	org   domain.Organization

	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger

	remoteList        remotelist.Model
	accountList       []domain.Account
	selectedAccountID domain.AccountID
	width             int
	height            int
}

// New creates a new account select model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	role string,
	org domain.Organization,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) Model {
	delegate := accountDelegate{theme: theme}

	return Model{
		ctx:        ctx,
		theme:      theme,
		role:       role,
		org:        org,
		services:   services,
		prefs:      prefs,
		logger:     logger,
		remoteList: remotelist.New(theme, delegate, "Loading accounts", logger),
		width:      80,
	}
}

// Init starts loading accounts.
func (m Model) Init() tea.Cmd {
	_, cmd := m.remoteList.InitWithLoader(m.loadAccounts())
	return cmd
}

// loadAccounts returns a command that loads accounts.
func (m Model) loadAccounts() tea.Cmd {
	return func() tea.Msg {
		m.logger.Info("loading accounts", "organizationID", m.org.ID)

		accounts, err := m.services.Accounts.List(m.ctx, m.org.ID)
		if err != nil {
			m.logger.Error("failed to load accounts", "error", err)
			return remotelist.LoadResultMsg{Items: nil, Err: err}
		}

		m.logger.Info("accounts loaded", "count", len(accounts))

		items := make([]list.Item, len(accounts))
		for i, acc := range accounts {
			items[i] = AccountListItem{Account: acc}
		}

		return remotelist.LoadResultMsg{Items: items, Err: nil}
	}
}

// selectAccount handles account selection.
func (m Model) selectAccount(accountID domain.AccountID) Model {
	m.selectedAccountID = accountID
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case remotelist.LoadResultMsg:
		if msg.Err == nil {
			m.accountList = make([]domain.Account, 0, len(msg.Items))
			for _, item := range msg.Items {
				if accItem, ok := item.(AccountListItem); ok {
					m.accountList = append(m.accountList, accItem.Account)
				}
			}

			userPref := m.prefs.GetDefaultAccountID()

			// No accounts → auto-select create
			if len(m.accountList) == 0 {
				m.logger.Debug("auto-selected create account", "reason", "no accounts found")
				m.selectedAccountID = createNewAccountID
			}

			// Has preference and exists → auto-select
			if userPref != "" {
				for _, acc := range m.accountList {
					if acc.ID == userPref {
						m.logger.Debug("auto-selected account from preference", "id", userPref, "name", acc.Name)
						m.services.SetAccountID(acc.ID)
						m = m.selectAccount(acc.ID)
						// Emit AccountSelectedMsg to trigger sync
						selectedAcc := acc
						cmds = append(cmds, func() tea.Msg {
							return AccountSelectedMsg{Organization: m.org, Account: selectedAcc}
						})
						break
					}
				}
			}

			// No preference and only 1 account → auto-select and save
			if userPref == "" && len(m.accountList) == 1 {
				acc := m.accountList[0]
				m.logger.Info("auto-selected account", "id", acc.ID, "name", acc.Name, "reason", "only one available")
				if err := m.prefs.SetDefaultAccountID(acc.ID); err != nil {
					m.logger.Error("failed to save account preference", "error", err)
				}
				m.services.SetAccountID(acc.ID)
				m = m.selectAccount(acc.ID)
				// Emit AccountSelectedMsg to trigger sync
				cmds = append(cmds, func() tea.Msg {
					return AccountSelectedMsg{Organization: m.org, Account: acc}
				})
			}
		}

	case tea.KeyPressMsg:
		switch msg.String() {
		case "r":
			if m.remoteList.HasError() {
				m.logger.Info("user requested retry")
				var cmd tea.Cmd
				m.remoteList, cmd = m.remoteList.Retry()
				cmds = append(cmds, cmd)
			}
		case "enter":
			if !m.remoteList.IsLoaded() {
				break
			}
			selected := m.remoteList.SelectedItem()
			if accItem, ok := selected.(AccountListItem); ok {
				m.logger.Info("account selected", "id", accItem.ID, "name", accItem.Name)
				if err := m.prefs.SetDefaultAccountID(accItem.ID); err != nil {
					m.logger.Error("failed to save account preference", "error", err)
				}
				m.services.SetAccountID(accItem.ID)
				m = m.selectAccount(accItem.ID)
				// Emit AccountSelectedMsg to trigger sync
				acc := accItem.Account
				cmds = append(cmds, func() tea.Msg {
					return AccountSelectedMsg{Organization: m.org, Account: acc}
				})
			}
		case "n":
			if !m.remoteList.IsLoaded() {
				break
			}
			m.logger.Info("user chose to create new account")
			m = m.selectAccount(createNewAccountID)
		}
	}

	var cmd tea.Cmd
	m.remoteList, cmd = m.remoteList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the account selection UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles

	if m.remoteList.IsBusy() {
		return m.remoteList.View()
	}

	if m.remoteList.HasError() {
		return m.remoteList.View()
	}

	title := themeStyles.Title.Render("Select your Datadog account")
	subtitle := themeStyles.Help.Render("Connect a Datadog account to sync metrics")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		m.remoteList.View(),
	)
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	m.remoteList = m.remoteList.SetWidth(width)
	return m
}

// IsBusy returns true while loading.
func (m Model) IsBusy() bool {
	return !m.remoteList.IsLoaded()
}

// HasError returns true if there's an error.
func (m Model) HasError() bool {
	return m.remoteList.HasError()
}

// Error returns the current error.
func (m Model) Error() error {
	return m.remoteList.Error()
}

// Help returns the key bindings for this step.
func (m Model) Help() help.KeyMap {
	if m.remoteList.IsBusy() {
		return keymap.Simple{Keys: []key.Binding{}}
	}

	if m.remoteList.HasError() {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}

	listKeys := m.remoteList.KeyMap()
	return keymap.Simple{
		Keys: []key.Binding{
			listKeys.CursorUp,
			listKeys.CursorDown,
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "create new"),
			),
		},
	}
}

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.selectedAccountID == "" {
		return nil, step.ErrNotReady
	}

	if m.selectedAccountID == createNewAccountID {
		return accountcreate.New(
			m.ctx,
			m.theme,
			m.role,
			m.org,
			m.services,
			m.prefs,
			m.logger,
		), nil
	}

	var selectedAccount domain.Account
	for _, acc := range m.accountList {
		if acc.ID == m.selectedAccountID {
			selectedAccount = acc
			break
		}
	}

	return datadogcheck.New(
		m.ctx,
		m.theme,
		m.role,
		m.org,
		selectedAccount,
		m.services,
		m.prefs,
		m.logger,
	), nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
