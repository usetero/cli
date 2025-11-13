package account

import (
	"context"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components/input"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
)

const cursorMarker = "┃"

// AccountCreator creates accounts
type AccountCreator interface {
	Create(ctx context.Context, orgID string, name string) (*api.Account, error)
}

// CreateStep handles creating a new account
type CreateStep struct {
	// Accumulated state from previous steps
	role  string
	orgID string

	// Services (defined by consumer interfaces)
	accountCreator      AccountCreator
	defaultAccountSaver DefaultAccountSaver

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
func NewCreateStep(role string, orgID string, accountCreator AccountCreator, defaultAccountSaver DefaultAccountSaver, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if accountCreator == nil {
		panic("accountCreator cannot be nil")
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

	inp := input.New(logger)
	inp.SetPlaceholder("Production")
	inp.SetCharLimit(100)

	return &CreateStep{
		role:                role,
		orgID:               orgID,
		accountCreator:      accountCreator,
		defaultAccountSaver: defaultAccountSaver,
		apiClient:           apiClient,
		logger:              logger,
		input:               inp,
		width:               80,
		globalBindings:      globalBindings,
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

		// Save account to preferences
		if err := s.defaultAccountSaver.SetDefaultAccountID(msg.account.ID); err != nil {
			s.logger.Error("failed to save account preference", "error", err)
			s.err = err
			return s, nil
		}
		s.logger.Debug("account saved to preferences", "accountID", msg.account.ID)

		// Clear any previous error and mark success
		s.err = nil
		s.created = true
		s.createdAccount = msg.account
		return s, nil
	}

	return s, cmd
}

// createAccount creates a new account via the API
func (s *CreateStep) createAccount(name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		s.logger.Info("creating account", "name", name, "organizationID", s.orgID)

		account, err := s.accountCreator.Create(ctx, s.orgID, name)
		if err != nil {
			return createAccountMsg{err: err}
		}

		return createAccountMsg{account: account}
	}
}

// View renders the create account UI
func (s *CreateStep) View() string {
<<<<<<< HEAD
	theme := styles.CurrentTheme()

	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)
=======
	common := styles.Common()
>>>>>>> 17e8dd9 (chore: initial commit)

	if s.creating {
		return lipgloss.JoinVertical(
			lipgloss.Left,
<<<<<<< HEAD
			titleStyle.Render("Creating account..."),
		)
	}

	title := titleStyle.Render("Create a new account")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(theme.TextSubtle)
	subtitle := subtitleStyle.Render("Enter your account name")
=======
			common.Title.Render("Creating account..."),
		)
	}

	title := common.Title.Render("Create a new account")
	subtitle := common.Subtitle.Render("Enter your account name")
>>>>>>> 17e8dd9 (chore: initial commit)

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

<<<<<<< HEAD
	helpStyle := lipgloss.NewStyle().
		Foreground(theme.TextMuted)
	help := helpStyle.Render("This groups your observability tools and services")
=======
	help := common.Help.Render("This groups your observability tools and services")
>>>>>>> 17e8dd9 (chore: initial commit)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		subtitle,
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

// IsComplete returns true if the account has been created successfully
func (s *CreateStep) IsComplete() bool {
	return s.created && s.err == nil
}

// CreatedAccountID returns the ID of the created account
func (s *CreateStep) CreatedAccountID() string {
	if s.createdAccount != nil {
		return s.createdAccount.ID
	}
	return ""
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
func (s *CreateStep) Next() step.Step {
	// Create Datadog service for next step
	datadogService := api.NewDatadogAccountService(s.apiClient, s.logger)

	// Check for Datadog with accumulated data
	return datadog.NewCheckDatadogStep(s.role, s.orgID, s.CreatedAccountID(), datadogService, s.apiClient, s.logger, s.globalBindings)
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
