package datadog

import (
	"context"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/services"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
)

// DatadogAccountChecker checks if an account has a Datadog integration
type DatadogAccountChecker interface {
	HasAccount(ctx context.Context, accountID string) (bool, error)
	GetAccount(ctx context.Context, accountID string) (*api.DatadogAccount, error)
}

// CheckDatadogStep checks if the account has a Datadog integration configured
type CheckDatadogStep struct {
	// Accumulated state from previous steps
	role      string
	orgID     string
	accountID string

	// Services (defined by consumer interfaces)
	datadogChecker DatadogAccountChecker

	// Pass-through to next step
	apiClient api.Client
	logger    log.Logger

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
func NewCheckDatadogStep(role string, orgID string, accountID string, datadogChecker DatadogAccountChecker, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if datadogChecker == nil {
		panic("datadogChecker cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	return &CheckDatadogStep{
		role:           role,
		orgID:          orgID,
		accountID:      accountID,
		datadogChecker: datadogChecker,
		apiClient:      apiClient,
		logger:         logger,
		width:          80,
		globalBindings: globalBindings,
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
		ctx := context.Background()
		s.logger.Info("checking datadog account", log.String("accountID", s.accountID))

		hasDatadog, err := s.datadogChecker.HasAccount(ctx, s.accountID)
		if err != nil {
			return checkDatadogMsg{err: err}
		}

		if !hasDatadog {
			return checkDatadogMsg{hasDatadog: false}
		}

		// Fetch the Datadog account details
		account, err := s.datadogChecker.GetAccount(ctx, s.accountID)
		if err != nil {
			return checkDatadogMsg{err: err}
		}

		return checkDatadogMsg{hasDatadog: true, datadogAccount: account}
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
	common := styles.Common()

	if s.checking {
		return common.Title.Render("Checking Datadog account...")
	}

	if s.hasDatadog {
		return common.Success.Render("✓ Datadog account found")
	}

	return common.Title.Render("No Datadog account found")
}

// SetSize sets the width available for rendering
func (s *CheckDatadogStep) SetSize(width, height int) {
	s.width = width
}

// IsComplete returns true when check is done successfully
func (s *CheckDatadogStep) IsComplete() bool {
	return s.checked && s.err == nil
}

// NeedsDatadogSetup returns true if account doesn't have Datadog configured
func (s *CheckDatadogStep) NeedsDatadogSetup() bool {
	return s.checked && !s.hasDatadog
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
func (s *CheckDatadogStep) Next() step.Step {
	// Conditional branching based on whether account has Datadog
	if s.NeedsDatadogSetup() {
		// No Datadog account - go to Datadog setup flow
		return NewSelectRegionStep(s.role, s.orgID, s.accountID, s.apiClient, s.logger, s.globalBindings)
	}

	// Create service service for next step
	serviceService := api.NewServiceService(s.apiClient, s.logger)

	// Datadog account exists - go to service discovery
	datadogAccountID := s.datadogAccount.ID
	return services.NewDiscoveryStep(s.role, s.orgID, s.accountID, &datadogAccountID, serviceService, s.apiClient, s.logger, s.globalBindings)
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
