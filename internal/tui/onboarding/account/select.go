package account

import (
	"context"
	"fmt"
	"io"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/remotelist"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// AccountSelectedMsg is sent when an account is selected during onboarding.
// This triggers PowerSync to start syncing in the background.
type AccountSelectedMsg struct {
	Account api.Account
}

const createNewAccountID = "__CREATE_NEW__"

// AccountLister lists accounts
type AccountLister interface {
	List(ctx context.Context, orgID string) ([]api.Account, error)
}

// AccountItem implements list.Item for the list component.
// Exported for testing.
type AccountItem struct {
	ID   string
	Name string
}

func (i AccountItem) FilterValue() string { return i.Name }

// accountDelegate renders each account in the list
type accountDelegate struct {
	theme *styles.Theme
}

func (d accountDelegate) Height() int                             { return 1 }
func (d accountDelegate) Spacing() int                            { return 0 }
func (d accountDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d accountDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(AccountItem)
	if !ok {
		return
	}

	colors := d.theme.Colors

	str := i.Name
	if index == m.Index() {
		fn := lipgloss.NewStyle().
			Foreground(colors.Accent).
			Bold(true).
			Render
		_, _ = fmt.Fprint(w, fn("> "+str))
	} else {
		fn := lipgloss.NewStyle().
			Foreground(colors.Page.Text).
			Render
		_, _ = fmt.Fprint(w, fn("  "+str))
	}
}

// SelectStep handles selecting an account or choosing to create one.
type SelectStep struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Theme for styling
	theme *styles.Theme

	// Accumulated state from previous steps
	role string
	org  api.Organization

	// Services (defined by consumer interfaces)
	accountLister       AccountLister
	defaultAccountSaver DefaultAccountSaver

	// Pass-through to next step
	apiClient api.Client
	logger    log.Logger

	// UI state
	remoteList        *remotelist.Component
	accountsList      []api.Account
	selectedAccountID string
	selectedAccount   api.Account
	width             int
	globalBindings    []key.Binding
}

// NewSelectStep creates a new account selection step for the given organization
func NewSelectStep(ctx context.Context, theme *styles.Theme, role string, org api.Organization, accountLister AccountLister, defaultAccountSaver DefaultAccountSaver, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if accountLister == nil {
		panic("accountLister cannot be nil")
	}
	if defaultAccountSaver == nil {
		panic("defaultAccountSaver cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	delegate := accountDelegate{theme: theme}
	remoteList := remotelist.New(theme, delegate, "Loading accounts", logger)

	return &SelectStep{
		ctx:                 ctx,
		theme:               theme,
		role:                role,
		org:                 org,
		accountLister:       accountLister,
		defaultAccountSaver: defaultAccountSaver,
		apiClient:           apiClient,
		logger:              logger,
		remoteList:          remoteList,
		width:               80,
		globalBindings:      globalBindings,
	}
}

// Init starts loading accounts for the specified organization
func (s *SelectStep) Init() tea.Cmd {
	return s.remoteList.InitWithLoader(func() tea.Msg {
		s.logger.Info("loading accounts", "organizationID", s.org.ID)
		accounts, err := s.accountLister.List(s.ctx, s.org.ID)
		if err != nil {
			s.logger.Error("failed to load accounts", "error", err, "organizationID", s.org.ID)
			return remotelist.LoadResultMsg{Items: nil, Err: err}
		}

		s.logger.Info("accounts loaded", log.Int("count", len(accounts)))

		// Build list items from accounts
		items := make([]list.Item, len(accounts))
		for i, account := range accounts {
			items[i] = AccountItem{ID: account.ID, Name: account.Name}
		}

		return remotelist.LoadResultMsg{Items: items, Err: nil}
	})
}

// Update handles messages
func (s *SelectStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle remotelist's LoadResultMsg to apply auto-selection logic
	switch msg := msg.(type) {
	case remotelist.LoadResultMsg:
		if msg.Err == nil {
			// Extract accounts from items
			s.accountsList = make([]api.Account, 0, len(msg.Items))
			for _, item := range msg.Items {
				if ai, ok := item.(AccountItem); ok {
					s.accountsList = append(s.accountsList, api.Account{ID: ai.ID, Name: ai.Name})
				}
			}

			// Apply auto-selection logic
			userPref := s.defaultAccountSaver.GetDefaultAccountID()

			// Case 1: No accounts → auto-select "create"
			if len(s.accountsList) == 0 {
				s.selectedAccountID = createNewAccountID
				s.logger.Debug("auto-selected create account", "reason", "no accounts found")
			}

			// Case 2: Has preference AND exists → auto-select
			if userPref != "" {
				for _, account := range s.accountsList {
					if account.ID == userPref {
						s.selectedAccountID = userPref
						s.selectedAccount = account
						s.apiClient.SetAccountID(userPref)
						s.logger.Debug("auto-selected account from preference", "id", userPref, "name", account.Name)
						// Emit AccountSelectedMsg to trigger sync
						cmds = append(cmds, func() tea.Msg {
							return AccountSelectedMsg{Account: account}
						})
					}
				}
			}

			// Case 3: No preference AND only 1 account → auto-select and save
			if userPref == "" && len(s.accountsList) == 1 {
				account := s.accountsList[0]
				s.selectedAccountID = account.ID
				s.selectedAccount = account
				s.apiClient.SetAccountID(account.ID)
				if err := s.defaultAccountSaver.SetDefaultAccountID(account.ID); err != nil {
					s.logger.Error("failed to save account preference", "error", err)
				} else {
					s.logger.Debug("auto-selected account", "id", account.ID, "name", account.Name, "reason", "only one available")
				}
				// Emit AccountSelectedMsg to trigger sync
				cmds = append(cmds, func() tea.Msg {
					return AccountSelectedMsg{Account: account}
				})
			}

			// Case 4: User must select from list (selectedAccountID remains empty)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Retry loading if there's an error
			if s.remoteList.HasError() {
				s.logger.Info("user requested retry")
				cmd := s.remoteList.Retry()
				cmds = append(cmds, cmd)
			}
		case "enter":
			if !s.remoteList.IsLoaded() {
				break
			}
			selected := s.remoteList.SelectedItem()
			if ai, ok := selected.(AccountItem); ok {
				account := api.Account{ID: ai.ID, Name: ai.Name}
				s.selectedAccountID = ai.ID
				s.selectedAccount = account
				s.apiClient.SetAccountID(ai.ID)
				s.logger.Info("account selected", "id", ai.ID, "name", ai.Name)
				if err := s.defaultAccountSaver.SetDefaultAccountID(ai.ID); err != nil {
					s.logger.Error("failed to save account preference", "error", err)
				}
				// Emit AccountSelectedMsg to trigger sync
				cmds = append(cmds, func() tea.Msg {
					return AccountSelectedMsg{Account: account}
				})
			}
		case "n":
			if !s.remoteList.IsLoaded() {
				break
			}
			// User pressed 'n' to create new account
			s.selectedAccountID = createNewAccountID
			s.logger.Info("user chose to create new account")
		}
	}

	// Update remote list (handles loading, error, and list navigation)
	cmd := s.remoteList.Update(msg)
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

// View renders the account selection UI
func (s *SelectStep) View() string {
	// If still loading or has error, just show the remotelist (loader or empty)
	if s.remoteList.IsBusy() || s.remoteList.HasError() {
		return s.remoteList.View()
	}

	themeStyles := s.theme.Styles

	title := themeStyles.Title.Render("Select your account")
	subtitle := themeStyles.Subtitle.Render("This groups your observability tools and services")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		s.remoteList.View(),
	)
}

// SetSize sets the width available for rendering
func (s *SelectStep) SetSize(width, height int) {
	s.width = width
	s.remoteList.SetWidth(width)
}

// IsBusy returns true while loading accounts
func (s *SelectStep) IsBusy() bool {
	return !s.remoteList.IsLoaded()
}

// HasError returns true if the remotelist has an error
func (s *SelectStep) HasError() bool {
	return s.remoteList.HasError()
}

// Error returns the remotelist's error, or nil if no error
func (s *SelectStep) Error() error {
	return s.remoteList.Error()
}

// Next returns the next step after account selection
func (s *SelectStep) Next() (step.Step, error) {
	if s.selectedAccountID == "" {
		return nil, step.ErrNotReady
	}

	// User wants to create new account
	if s.selectedAccountID == createNewAccountID {
		accountService := api.NewAccountService(s.apiClient, s.logger)
		return NewCreateStep(s.ctx, s.theme, s.role, s.org, accountService, s.defaultAccountSaver, s.apiClient, s.logger, s.globalBindings), nil
	}

	// Create Datadog service for next step
	datadogService := api.NewDatadogAccountService(s.apiClient, s.logger)

	// User selected existing account, check for Datadog
	return datadog.NewCheckDatadogStep(s.ctx, s.theme, s.role, s.org, s.selectedAccount, datadogService, s.apiClient, s.logger, s.globalBindings), nil
}

// Help returns the key bindings for this step
func (s *SelectStep) Help() help.KeyMap {
	// If loading, no keybindings
	if s.remoteList.IsBusy() {
		return keymap.Simple{Keys: []key.Binding{}}
	}

	// If error, show retry keybinding
	if s.remoteList.HasError() {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}

	// Normal state: show list navigation and actions
	listKeys := s.remoteList.KeyMap()
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
