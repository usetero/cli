package organization

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

// OrganizationCreator creates organizations
type OrganizationCreator interface {
	Create(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error)
}

// DefaultAccountSaver defines the interface for saving default account preferences.
// This is needed because organization bootstrap also creates an account.
type DefaultAccountSaver interface {
	GetDefaultAccountID() string
	SetDefaultAccountID(accountID string) error
}

// CreateStep handles creating a new organization
type CreateStep struct {
	// Accumulated state from previous steps
	role string

	// Services (defined by consumer interfaces)
	organizationCreator OrganizationCreator
	defaultOrgSaver     DefaultOrgSaver
	defaultAccountSaver DefaultAccountSaver

	// Pass-through to next step
	apiClient api.Client
	logger    log.Logger

	// UI state
	input          *input.Component
	creating       bool
	created        bool
	createdResult  *api.OrganizationBootstrapResult
	err            error
	width          int
	globalBindings []key.Binding
}

// NewCreateStep creates a new organization creation step
func NewCreateStep(role string, organizationCreator OrganizationCreator, defaultOrgSaver DefaultOrgSaver, defaultAccountSaver DefaultAccountSaver, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if organizationCreator == nil {
		panic("organizationCreator cannot be nil")
	}
	if defaultOrgSaver == nil {
		panic("defaultOrgSaver cannot be nil")
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
	inp.SetPlaceholder("Acme Inc.")
	inp.SetCharLimit(100)

	return &CreateStep{
		role:                role,
		organizationCreator: organizationCreator,
		defaultOrgSaver:     defaultOrgSaver,
		defaultAccountSaver: defaultAccountSaver,
		apiClient:           apiClient,
		logger:              logger,
		input:               inp,
		width:               80,
		globalBindings:      globalBindings,
	}
}

// createOrgMsg is sent when org creation completes
type createOrgMsg struct {
	result *api.OrganizationBootstrapResult
	err    error
}

// Init focuses the input
func (s *CreateStep) Init() tea.Cmd {
	return nil // Input is already focused in constructor
}

// Update handles messages
func (s *CreateStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	// Always update input for cursor blinking
	inputCmd := s.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Retry on error
			if s.err != nil {
				s.err = nil
				s.creating = false
				return s, inputCmd
			}

			// Submit if not already creating
			if !s.creating && !s.created {
				name := s.input.Value()
				if name != "" {
					s.creating = true
					return s, tea.Batch(inputCmd, s.createOrganization(name))
				}
			}
		}

	case createOrgMsg:
		s.creating = false
		if msg.err != nil {
			s.logger.Error("failed to create organization", "error", msg.err)
			s.err = msg.err
			return s, inputCmd
		}

		s.logger.Info("organization created", "id", msg.result.Organization.ID, "name", msg.result.Organization.Name, "accountID", msg.result.Account.ID)

		// Save organization to preferences
		if err := s.defaultOrgSaver.SetDefaultOrgID(msg.result.Organization.ID); err != nil {
			s.logger.Error("failed to save org preference", "error", err)
			s.err = err
			return s, inputCmd
		}
		s.logger.Debug("organization saved to preferences", "orgID", msg.result.Organization.ID)

		// Organization bootstrap also creates an account - save it too
		if err := s.defaultAccountSaver.SetDefaultAccountID(msg.result.Account.ID); err != nil {
			s.logger.Error("failed to save account preference", "error", err)
			s.err = err
			return s, inputCmd
		}
		s.logger.Debug("account saved to preferences", "accountID", msg.result.Account.ID)

		// Clear any previous error and mark success
		s.err = nil
		s.created = true
		s.createdResult = msg.result
		return s, inputCmd
	}

	return s, inputCmd
}

// createOrganization creates a new organization via the API
func (s *CreateStep) createOrganization(name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		s.logger.Info("creating organization", log.String("name", name))

		result, err := s.organizationCreator.Create(ctx, name)
		if err != nil {
			return createOrgMsg{err: err}
		}

		return createOrgMsg{result: result}
	}
}

// View renders the create organization UI
func (s *CreateStep) View() string {
	common := styles.Common()

	if s.creating {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			common.Title.Render("Creating organization..."),
		)
	}

	title := common.Title.Render("Create a new organization")
	subtitle := common.Subtitle.Render("Enter your organization name")
	help := common.Help.Render("This will be your default workspace")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		subtitle,
		"",
		s.input.View(),
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

// IsComplete returns true if the organization has been created successfully
func (s *CreateStep) IsComplete() bool {
	return s.created && s.err == nil
}

// CreatedOrgID returns the ID of the created organization
func (s *CreateStep) CreatedOrgID() string {
	if s.createdResult != nil {
		return s.createdResult.Organization.ID
	}
	return ""
}

// CreatedAccountID returns the ID of the created account
func (s *CreateStep) CreatedAccountID() string {
	if s.createdResult != nil && s.createdResult.Account != nil {
		return s.createdResult.Account.ID
	}
	return ""
}

// IsBusy returns true while creating the organization
func (s *CreateStep) IsBusy() bool {
	return s.creating
}

// HasError returns true if there was an error creating the organization
func (s *CreateStep) HasError() bool {
	return s.err != nil
}

// Error returns the current error, or nil if no error
func (s *CreateStep) Error() error {
	return s.err
}

// Next returns the next step after creating organization
func (s *CreateStep) Next() step.Step {
	// Skip account selection since bootstrap creates it automatically
	// Go to Datadog region selection
	return datadog.NewSelectRegionStep(s.role, s.CreatedOrgID(), s.CreatedAccountID(), s.apiClient, s.logger, s.globalBindings)
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
