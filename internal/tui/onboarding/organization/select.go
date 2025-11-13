package organization

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/remotelist"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/account"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
)

const createNewOrgID = "__CREATE_NEW__"

// OrganizationLister lists organizations
type OrganizationLister interface {
	List(ctx context.Context) ([]api.Organization, error)
}

// orgItem implements list.Item for the list component
type orgItem struct {
	id   string
	name string
}

func (i orgItem) FilterValue() string { return i.name }

// orgDelegate renders each organization in the list
type orgDelegate struct{}

func (d orgDelegate) Height() int                             { return 1 }
func (d orgDelegate) Spacing() int                            { return 0 }
func (d orgDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d orgDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(orgItem)
	if !ok {
		return
	}

	theme := styles.CurrentTheme()

	str := i.name
	if index == m.Index() {
		fn := lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true).
			Render
		_, _ = fmt.Fprint(w, fn("> "+str))
	} else {
		fn := lipgloss.NewStyle().
			Foreground(theme.Text).
			Render
		_, _ = fmt.Fprint(w, fn("  "+str))
	}
}

// SelectStep handles selecting an organization or choosing to create one.
type SelectStep struct {
	// Accumulated state from previous steps
	role string

	// Services (defined by consumer interfaces)
	organizationLister  OrganizationLister
	defaultOrgSaver     DefaultOrgSaver
	defaultAccountSaver DefaultAccountSaver

	// Pass-through to next step
	apiClient api.Client
	logger    log.Logger

	// UI state
	remoteList     *remotelist.Component
	orgs           []api.Organization
	selectedOrgID  string
	width          int
	globalBindings []key.Binding
}

// NewSelectStep creates a new organization selection step
func NewSelectStep(role string, organizationLister OrganizationLister, apiClient api.Client, defaultOrgSaver DefaultOrgSaver, defaultAccountSaver DefaultAccountSaver, logger log.Logger, globalBindings []key.Binding) step.Step {
	if organizationLister == nil {
		panic("organizationLister cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if defaultOrgSaver == nil {
		panic("defaultOrgSaver cannot be nil")
	}
	if defaultAccountSaver == nil {
		panic("defaultAccountSaver cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	delegate := orgDelegate{}
	remoteList := remotelist.New(delegate, "Loading organizations", logger)

	return &SelectStep{
		role:                role,
		organizationLister:  organizationLister,
		defaultOrgSaver:     defaultOrgSaver,
		defaultAccountSaver: defaultAccountSaver,
		apiClient:           apiClient,
		logger:              logger,
		remoteList:          remoteList,
		width:               80,
		globalBindings:      globalBindings,
	}
}

// Init starts loading organizations
func (s *SelectStep) Init() tea.Cmd {
	return s.remoteList.InitWithLoader(func() tea.Msg {
		s.logger.Info("loading organizations")
		ctx := context.Background()

		orgs, err := s.organizationLister.List(ctx)
		if err != nil {
			s.logger.Error("failed to load organizations", "error", err)
			return remotelist.LoadResultMsg{Items: nil, Err: err}
		}

		s.logger.Info("organizations loaded", log.Int("count", len(orgs)))

		// Build list items from orgs
		items := make([]list.Item, len(orgs))
		for i, org := range orgs {
			items[i] = orgItem{id: org.ID, name: org.Name}
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
			// Extract orgs from items
			s.orgs = make([]api.Organization, 0, len(msg.Items))
			for _, item := range msg.Items {
				if orgItem, ok := item.(orgItem); ok {
					s.orgs = append(s.orgs, api.Organization{ID: orgItem.id, Name: orgItem.name})
				}
			}

			// Apply auto-selection logic
			userPref := s.defaultOrgSaver.GetDefaultOrgID()

			// Case 1: No orgs → auto-select "create" to fast-forward
			if len(s.orgs) == 0 {
				s.selectedOrgID = createNewOrgID
				s.logger.Debug("auto-selected create organization", "reason", "no organizations found")
			}

			// Case 2: Has preference AND exists → auto-select
			if userPref != "" {
				for _, org := range s.orgs {
					if org.ID == userPref {
						s.selectedOrgID = userPref
						s.logger.Debug("auto-selected organization from preference", "id", userPref, "name", org.Name)
					}
				}
			}

			// Case 3: No preference AND only 1 org → auto-select and save
			if userPref == "" && len(s.orgs) == 1 {
				s.selectedOrgID = s.orgs[0].ID
				s.logger.Info("auto-selected organization", "id", s.orgs[0].ID, "name", s.orgs[0].Name, "reason", "only one available")
				if err := s.defaultOrgSaver.SetDefaultOrgID(s.orgs[0].ID); err != nil {
					s.logger.Error("failed to save organization preference", "error", err)
				}
			}

			// Case 4: User must select from list (selectedOrgID remains empty)
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
			if org, ok := selected.(orgItem); ok {
				s.selectedOrgID = org.id
				s.logger.Info("organization selected", "id", org.id, "name", org.name)
				if err := s.defaultOrgSaver.SetDefaultOrgID(org.id); err != nil {
					s.logger.Error("failed to save organization preference", "error", err)
				}
			}
		case "n":
			if !s.remoteList.IsLoaded() {
				break
			}
			// User pressed 'n' to create new organization
			s.selectedOrgID = createNewOrgID
			s.logger.Info("user chose to create new organization")
		}
	}

	// Update remote list (handles loading, error, and list navigation)
	cmd := s.remoteList.Update(msg)
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

// View renders the organization selection UI
func (s *SelectStep) View() string {
	// If still loading or has error, just show the remotelist (loader or empty)
	if s.remoteList.IsBusy() || s.remoteList.HasError() {
		return s.remoteList.View()
	}

	common := styles.Common()

	title := common.Title.Render("Select your organization")
	subtitle := common.Subtitle.Render("This will be your default workspace")

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

// IsComplete returns true if an organization has been selected
func (s *SelectStep) IsComplete() bool {
	return s.selectedOrgID != ""
}

// IsCreateSelected returns true if user chose to create a new organization
func (s *SelectStep) IsCreateSelected() bool {
	return s.selectedOrgID == createNewOrgID
}

// SelectedOrgID returns the ID of the selected organization (excluding "create new")
func (s *SelectStep) SelectedOrgID() string {
	if s.selectedOrgID == createNewOrgID {
		return ""
	}
	return s.selectedOrgID
}

// IsBusy returns true while loading organizations
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

// Next returns the next step after organization selection
func (s *SelectStep) Next() step.Step {
	// Conditional branching based on user's selection
	if s.IsCreateSelected() {
		// Create organization service for next step
		organizationService := api.NewOrganizationService(s.apiClient, s.logger)

		// User wants to create new org - pass role forward
		return NewCreateStep(s.role, organizationService, s.defaultOrgSaver, s.defaultAccountSaver, s.apiClient, s.logger, s.globalBindings)
	}

	// Create account service for next step
	accountService := api.NewAccountService(s.apiClient, s.logger)

	// User selected existing org - pass role, orgID, and services forward
	return account.NewSelectStep(s.role, s.SelectedOrgID(), accountService, s.defaultAccountSaver, s.apiClient, s.logger, s.globalBindings)
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
