package datadog

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// CheckDatadogStep checks if the account has a Datadog integration configured
type CheckDatadogStep struct {
	// Context for API calls
	ctx context.Context

	// Theme for styling
	theme *styles.Theme

	// Accumulated state from previous steps
	role    string
	org     api.Organization
	account api.Account

	// Services
	datadogAccounts api.DatadogAccounts
	workspaces      api.Workspaces
	preferences     preferences.Preferences
	apiClient       api.Client
	logger          log.Logger

	// UI state
	checking       bool
	checked        bool
	hasDatadog     bool
	datadogAccount *api.DatadogAccount // The found Datadog account
	err            error
	width          int
	globalBindings []key.Binding
}

// NewCheckDatadogStep creates a new Datadog account check step
func NewCheckDatadogStep(ctx context.Context, theme *styles.Theme, role string, org api.Organization, account api.Account, datadogAccounts api.DatadogAccounts, workspaces api.Workspaces, prefs preferences.Preferences, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if datadogAccounts == nil {
		panic("datadogAccounts cannot be nil")
	}
	if workspaces == nil {
		panic("workspaces cannot be nil")
	}
	if prefs == nil {
		panic("preferences cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	return &CheckDatadogStep{
		ctx:             ctx,
		theme:           theme,
		role:            role,
		org:             org,
		account:         account,
		datadogAccounts: datadogAccounts,
		workspaces:      workspaces,
		preferences:     prefs,
		apiClient:       apiClient,
		logger:          logger,
		width:           80,
		globalBindings:  globalBindings,
	}
}

// checkDatadogMsg is sent when check completes
type checkDatadogMsg struct {
	hasDatadog     bool
	datadogAccount *api.DatadogAccount
	err            error
}

// Init starts checking for Datadog account
func (s *CheckDatadogStep) Init() tea.Cmd {
	s.checking = true
	return s.checkDatadogAccount()
}

// checkDatadogAccount checks if account has Datadog configured and fetches it
func (s *CheckDatadogStep) checkDatadogAccount() tea.Cmd {
	return func() tea.Msg {
		s.logger.Info("checking datadog account", log.String("accountID", s.account.ID))

		hasDatadog, err := s.datadogAccounts.HasAccount(s.ctx, s.account.ID)
		if err != nil {
			return checkDatadogMsg{err: err}
		}

		if !hasDatadog {
			return checkDatadogMsg{hasDatadog: false}
		}

		// Fetch the Datadog account details
		datadogAccount, err := s.datadogAccounts.GetAccount(s.ctx, s.account.ID)
		if err != nil {
			return checkDatadogMsg{err: err}
		}

		return checkDatadogMsg{hasDatadog: true, datadogAccount: datadogAccount}
	}
}

// Update handles messages
func (s *CheckDatadogStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case checkDatadogMsg:
		s.checking = false
		if msg.err != nil {
			s.logger.Error("failed to check datadog account", "error", msg.err)
			s.err = msg.err
			return s, nil
		}

		// Clear any previous error
		s.err = nil
		if msg.hasDatadog {
			s.logger.Info("datadog account found")
			s.datadogAccount = msg.datadogAccount
		} else {
			s.logger.Info("no datadog account found")
		}
		s.checked = true
		s.hasDatadog = msg.hasDatadog
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Retry checking Datadog account if there was an error
			if s.err != nil {
				s.logger.Info("retrying datadog account check")
				s.err = nil
				s.checking = true
				return s, s.checkDatadogAccount()
			}
		}
	}

	return s, nil
}

// View renders the check UI
func (s *CheckDatadogStep) View() string {
	themeStyles := s.theme.Styles

	if s.checking {
		return themeStyles.Title.Render("Checking Datadog account...")
	}

	if s.hasDatadog {
		return themeStyles.Success.Render("✓ Datadog account found")
	}

	return themeStyles.Title.Render("No Datadog account found")
}

// SetSize sets the width available for rendering
func (s *CheckDatadogStep) SetSize(width, height int) {
	s.width = width
}

// IsBusy returns true while checking
func (s *CheckDatadogStep) IsBusy() bool {
	return s.checking
}

// HasError returns true if there was an error checking Datadog account
func (s *CheckDatadogStep) HasError() bool {
	return s.err != nil
}

// Error returns the current error, or nil if no error
func (s *CheckDatadogStep) Error() error {
	return s.err
}

// Next returns the next step after checking Datadog account
func (s *CheckDatadogStep) Next() (step.Step, error) {
	if s.err != nil {
		return nil, s.err
	}
	if !s.checked {
		return nil, step.ErrNotReady
	}

	// Conditional branching based on whether account has Datadog
	if !s.hasDatadog {
		// No Datadog account - go to Datadog setup flow
		return NewSelectRegionStep(s.ctx, s.theme, s.role, s.org, s.account, s.workspaces, s.preferences, s.apiClient, s.logger, s.globalBindings), nil
	}

	// Create datadog account service for status polling
	datadogAccountService := api.NewDatadogAccountService(s.apiClient, s.logger)

	// Datadog account exists - go to unified discovery step
	datadogAccountID := s.datadogAccount.ID
	return NewDiscoveryStep(s.ctx, s.theme, s.role, s.org, s.account, &datadogAccountID, datadogAccountService, s.workspaces, s.preferences, s.logger, s.globalBindings), nil
}

// Help returns the key bindings for this step
func (s *CheckDatadogStep) Help() help.KeyMap {
	// Show retry option if there's an error
	if s.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}

	// No user interaction during normal checking
	return keymap.Simple{Keys: []key.Binding{}}
}

// Close releases any resources held by the step.
func (s *CheckDatadogStep) Close() error {
	return nil
}
