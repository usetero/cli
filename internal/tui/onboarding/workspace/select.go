package workspace

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
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/remotelist"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/onboarding/sync"
)

// WorkspaceSelectedMsg is sent when a workspace is selected during onboarding.
// This triggers PowerSync to start syncing in the background.
type WorkspaceSelectedMsg struct {
	Organization api.Organization
	Account      api.Account
	Workspace    api.Workspace
}

// WorkspaceItem implements list.Item for the list component.
type WorkspaceItem struct {
	ID   string
	Name string
}

func (i WorkspaceItem) FilterValue() string { return i.Name }

// workspaceDelegate renders each workspace in the list
type workspaceDelegate struct {
	theme *styles.Theme
}

func (d workspaceDelegate) Height() int                             { return 1 }
func (d workspaceDelegate) Spacing() int                            { return 0 }
func (d workspaceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d workspaceDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(WorkspaceItem)
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

// SelectStep handles selecting a workspace.
type SelectStep struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Theme for styling
	theme *styles.Theme

	// Accumulated state from previous steps
	org     api.Organization
	account api.Account

	// Services
	workspaces  api.Workspaces
	preferences preferences.Preferences

	// Pass-through to next step
	logger log.Logger

	// UI state
	remoteList          *remotelist.Component
	workspacesList      []api.Workspace
	selectedWorkspaceID string
	selectedWorkspace   api.Workspace
	width               int
	globalBindings      []key.Binding
}

// NewSelectStep creates a new workspace selection step for the given account
func NewSelectStep(
	ctx context.Context,
	theme *styles.Theme,
	org api.Organization,
	account api.Account,
	workspaces api.Workspaces,
	prefs preferences.Preferences,
	logger log.Logger,
	globalBindings []key.Binding,
) step.Step {
	if workspaces == nil {
		panic("workspaces cannot be nil")
	}
	if prefs == nil {
		panic("preferences cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	delegate := workspaceDelegate{theme: theme}
	remoteList := remotelist.New(theme, delegate, "Loading workspaces", logger)

	return &SelectStep{
		ctx:            ctx,
		theme:          theme,
		org:            org,
		account:        account,
		workspaces:     workspaces,
		preferences:    prefs,
		logger:         logger,
		remoteList:     remoteList,
		width:          80,
		globalBindings: globalBindings,
	}
}

// Init starts loading workspaces for the specified account
func (s *SelectStep) Init() tea.Cmd {
	return s.remoteList.InitWithLoader(func() tea.Msg {
		s.logger.Info("loading workspaces", "accountID", s.account.ID)
		workspaces, err := s.workspaces.List(s.ctx, s.account.ID)
		if err != nil {
			s.logger.Error("failed to load workspaces", "error", err, "accountID", s.account.ID)
			return remotelist.LoadResultMsg{Items: nil, Err: err}
		}

		s.logger.Info("workspaces loaded", log.Int("count", len(workspaces)))

		// Build list items from workspaces
		items := make([]list.Item, len(workspaces))
		for i, workspace := range workspaces {
			items[i] = WorkspaceItem{ID: workspace.ID, Name: workspace.Name}
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
			// Extract workspaces from items
			s.workspacesList = make([]api.Workspace, 0, len(msg.Items))
			for _, item := range msg.Items {
				if wi, ok := item.(WorkspaceItem); ok {
					s.workspacesList = append(s.workspacesList, api.Workspace{ID: wi.ID, Name: wi.Name})
				}
			}

			// Apply auto-selection logic
			userPref := s.preferences.GetDefaultWorkspaceID()

			// Case 1: Has preference AND exists → auto-select
			if userPref != "" {
				for _, workspace := range s.workspacesList {
					if workspace.ID == userPref {
						s.selectedWorkspaceID = userPref
						s.selectedWorkspace = workspace
						s.logger.Debug("auto-selected workspace from preference", "id", userPref, "name", workspace.Name)
						// Emit WorkspaceSelectedMsg to trigger sync
						selectedWorkspace := workspace
						selectedOrg := s.org
						selectedAccount := s.account
						cmds = append(cmds, func() tea.Msg {
							return WorkspaceSelectedMsg{Organization: selectedOrg, Account: selectedAccount, Workspace: selectedWorkspace}
						})
					}
				}
			}

			// Case 2: No preference AND only 1 workspace → auto-select and save
			if userPref == "" && len(s.workspacesList) == 1 {
				workspace := s.workspacesList[0]
				s.selectedWorkspaceID = workspace.ID
				s.selectedWorkspace = workspace
				if err := s.preferences.SetDefaultWorkspaceID(workspace.ID); err != nil {
					s.logger.Error("failed to save workspace preference", "error", err)
				} else {
					s.logger.Debug("auto-selected workspace", "id", workspace.ID, "name", workspace.Name, "reason", "only one available")
				}
				// Emit WorkspaceSelectedMsg to trigger sync
				selectedOrg := s.org
				selectedAccount := s.account
				cmds = append(cmds, func() tea.Msg {
					return WorkspaceSelectedMsg{Organization: selectedOrg, Account: selectedAccount, Workspace: workspace}
				})
			}

			// Case 3: User must select from list (selectedWorkspaceID remains empty)
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
			if wi, ok := selected.(WorkspaceItem); ok {
				workspace := api.Workspace{ID: wi.ID, Name: wi.Name}
				s.selectedWorkspaceID = wi.ID
				s.selectedWorkspace = workspace
				s.logger.Info("workspace selected", "id", wi.ID, "name", wi.Name)
				if err := s.preferences.SetDefaultWorkspaceID(wi.ID); err != nil {
					s.logger.Error("failed to save workspace preference", "error", err)
				}
				// Emit WorkspaceSelectedMsg to trigger sync
				selectedOrg := s.org
				selectedAccount := s.account
				cmds = append(cmds, func() tea.Msg {
					return WorkspaceSelectedMsg{Organization: selectedOrg, Account: selectedAccount, Workspace: workspace}
				})
			}
		}
	}

	// Update remote list (handles loading, error, and list navigation)
	cmd := s.remoteList.Update(msg)
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

// View renders the workspace selection UI
func (s *SelectStep) View() string {
	// If still loading or has error, just show the remotelist (loader or empty)
	if s.remoteList.IsBusy() || s.remoteList.HasError() {
		return s.remoteList.View()
	}

	themeStyles := s.theme.Styles

	title := themeStyles.Title.Render("Select your workspace")
	subtitle := themeStyles.Subtitle.Render("Workspaces organize your conversations and telemetry analysis")

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

// IsBusy returns true while loading workspaces
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

// Next returns the next step after workspace selection
func (s *SelectStep) Next() (step.Step, error) {
	if s.selectedWorkspaceID == "" {
		return nil, step.ErrNotReady
	}

	// Proceed to sync step
	return sync.New(s.ctx, s.theme, s.org, s.account, s.selectedWorkspace, s.logger, s.globalBindings), nil
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
		},
	}
}

// Close releases any resources held by the step.
func (s *SelectStep) Close() error {
	return nil
}

// Organization returns the organization.
func (s *SelectStep) Organization() api.Organization {
	return s.org
}

// Account returns the account.
func (s *SelectStep) Account() api.Account {
	return s.account
}

// Workspace returns the selected workspace.
func (s *SelectStep) Workspace() api.Workspace {
	return s.selectedWorkspace
}
