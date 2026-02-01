package account

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

const cursorMarker = "┃"

// CreateStep handles creating a new account
type CreateStep struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Theme for styling
	theme *styles.Theme

	// Accumulated state from previous steps
	role string
	org  api.Organization

	// Services
	accounts    api.Accounts
	preferences preferences.Preferences

	// Pass-through to next step
	apiClient api.Client
	logger    log.Logger

	// UI state
	input          *input.Component
	creating       bool
	created        bool
	createdAccount *api.Account
	err            error
	width          int
	globalBindings []key.Binding
}

// NewCreateStep creates a new account creation step for the given organization
func NewCreateStep(ctx context.Context, theme *styles.Theme, role string, org api.Organization, accounts api.Accounts, prefs preferences.Preferences, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if accounts == nil {
		panic("accounts cannot be nil")
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

	inp := input.New(theme, logger)
	inp.SetPlaceholder("Production")
	inp.SetCharLimit(100)

	return &CreateStep{
		ctx:            ctx,
		theme:          theme,
		role:           role,
		org:            org,
		accounts:       accounts,
		preferences:    prefs,
		apiClient:      apiClient,
		logger:         logger,
		input:          inp,
		width:          80,
		globalBindings: globalBindings,
	}
}

// createAccountMsg is sent when account creation completes
type createAccountMsg struct {
	account *api.Account
	err     error
}

// Init focuses the input
func (s *CreateStep) Init() tea.Cmd {
	return s.input.Focus()
}

// Update handles messages
func (s *CreateStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	// Always update input for cursor blinking
	cmd := s.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Retry on error
			if s.err != nil {
				s.err = nil
				s.creating = false
				return s, nil
			}

			// Submit if not already creating
			if !s.creating && !s.created {
				name := s.input.Value()
				if name != "" {
					s.creating = true
					return s, s.createAccount(name)
				}
			}
		}

	case createAccountMsg:
		s.creating = false
		if msg.err != nil {
			s.logger.Error("failed to create account", "error", msg.err)
			s.err = msg.err
			return s, nil
		}

		s.logger.Info("account created", "id", msg.account.ID, "name", msg.account.Name)

		// Set account ID on client for subsequent requests
		s.apiClient.SetAccountID(msg.account.ID)

		// Save account to preferences
		if err := s.preferences.SetDefaultAccountID(msg.account.ID); err != nil {
			s.logger.Error("failed to save account preference", "error", err)
			s.err = err
			return s, nil
		}
		s.logger.Debug("account saved to preferences", "accountID", msg.account.ID)

		// Clear any previous error and mark success
		s.err = nil
		s.created = true
		s.createdAccount = msg.account

		// Emit AccountSelectedMsg to trigger sync
		account := *msg.account
		org := s.org
		return s, func() tea.Msg {
			return AccountSelectedMsg{Organization: org, Account: account}
		}
	}

	return s, cmd
}

// createAccount creates a new account via the API
func (s *CreateStep) createAccount(name string) tea.Cmd {
	return func() tea.Msg {
		s.logger.Info("creating account", "name", name, "organizationID", s.org.ID)

		account, err := s.accounts.Create(s.ctx, s.org.ID, name)
		if err != nil {
			return createAccountMsg{err: err}
		}

		return createAccountMsg{account: account}
	}
}

// View renders the create account UI
func (s *CreateStep) View() string {
	themeStyles := s.theme.Styles

	if s.creating {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			themeStyles.Title.Render("Creating account..."),
		)
	}

	title := themeStyles.Title.Render("Create a new account")
	prompt := themeStyles.Body.Render("Enter your account name")

	// Input with cursor marker
	inputCursor := s.input.Cursor()
	inputView := s.input.View()
	var inputLine string
	if inputCursor != nil {
		if inputCursor.X <= len(inputView) {
			inputLine = inputView[:inputCursor.X] + cursorMarker + inputView[inputCursor.X:]
		} else {
			inputLine = inputView + cursorMarker
		}
	} else {
		inputLine = inputView
	}

	help := themeStyles.Help.Render("This groups your observability tools and services")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		prompt,
		"",
		inputLine,
		"",
		help,
	)

	return content
}

// SetSize sets the width available for rendering
func (s *CreateStep) SetSize(width, height int) {
	s.width = width
	if width > 10 {
		s.input.SetWidth(width - 6)
	}
}

// IsBusy returns true while creating the account
func (s *CreateStep) IsBusy() bool {
	return s.creating
}

// HasError returns true if account creation failed
func (s *CreateStep) HasError() bool {
	return s.err != nil
}

// Error returns the creation error, or nil if no error
func (s *CreateStep) Error() error {
	return s.err
}

// Next returns the next step after creating account
func (s *CreateStep) Next() (step.Step, error) {
	if s.err != nil {
		return nil, s.err
	}
	if !s.created || s.createdAccount == nil {
		return nil, step.ErrNotReady
	}

	// Create Datadog service for next step
	datadogService := api.NewDatadogAccountService(s.apiClient, s.logger)

	// Check for Datadog with accumulated data
	return datadog.NewCheckDatadogStep(s.ctx, s.theme, s.role, s.org, *s.createdAccount, datadogService, s.apiClient, s.logger, s.globalBindings), nil
}

// Help returns the key bindings for this step
func (s *CreateStep) Help() help.KeyMap {
	// Show retry hint if there's an error
	if s.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "retry"),
				),
			},
		}
	}

	// Normal state: show submit
	return keymap.Simple{
		Keys: []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "submit"),
			),
		},
	}
}
