package sync

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// Step waits for sync to complete before transitioning to the app.
// By the time the user reaches this step (after workspace selection),
// sync should already be done. This is a quick sanity check.
type Step struct {
	ctx   context.Context
	theme *styles.Theme

	org       api.Organization
	account   api.Account
	workspace api.Workspace

	logger         log.Logger
	globalBindings []key.Binding

	// UI state
	spinner    spinner.Model
	ready      bool
	width      int
	status     powersync.Status
	syncStatus *powersync.SyncStatus
	lastError  error
}

// New creates a new sync step.
func New(
	ctx context.Context,
	theme *styles.Theme,
	org api.Organization,
	account api.Account,
	workspace api.Workspace,
	logger log.Logger,
	globalBindings []key.Binding,
) step.Step {
	if logger == nil {
		panic("logger cannot be nil")
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	return &Step{
		ctx:            ctx,
		theme:          theme,
		org:            org,
		account:        account,
		workspace:      workspace,
		logger:         logger,
		globalBindings: globalBindings,
		spinner:        s,
		width:          80,
	}
}

// Init starts the spinner and queries sync status.
func (s *Step) Init() tea.Cmd {
	s.logger.Debug("sync step initialized, querying sync status")
	return tea.Batch(
		s.spinner.Tick,
		func() tea.Msg { return powersync.SyncStatusQueryMsg{} },
	)
}

// Update handles messages.
func (s *Step) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case powersync.SyncReadyMsg:
		s.ready = true
		s.logger.Info("sync ready, proceeding to app")
		return s, nil

	case powersync.StatusUpdateMsg:
		s.status = msg.Status
		s.syncStatus = msg.SyncStatus
		s.lastError = msg.LastError
		return s, nil
	}

	// Update spinner while waiting
	if !s.ready {
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
	}

	return s, nil
}

// View renders the sync waiting UI.
func (s *Step) View() string {
	themeStyles := s.theme.Styles

	title := themeStyles.Title.Render("Getting ready")

	if s.ready {
		statusMsg := themeStyles.Body.Render("Ready!")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusMsg)
	}

	// Build status message based on current state
	statusText := s.statusText()
	statusMsg := s.spinner.View() + " " + themeStyles.Body.Render(statusText)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", statusMsg)
}

// statusText returns human-readable status text.
func (s *Step) statusText() string {
	// Show error if present
	if s.lastError != nil {
		return fmt.Sprintf("Error: %v", s.lastError)
	}

	// Show download progress if available
	if s.syncStatus != nil && s.syncStatus.Downloading != nil {
		downloaded, total := s.syncStatus.Downloading.TotalProgress()
		if total > 0 {
			return fmt.Sprintf("Syncing your data... (%d/%d)", downloaded, total)
		}
	}

	// Show status-based text
	switch s.status {
	case powersync.StatusConnecting:
		return "Connecting..."
	case powersync.StatusSyncing:
		return "Syncing your data..."
	case powersync.StatusReconnecting:
		return "Reconnecting..."
	case powersync.StatusError:
		return "Error syncing"
	default:
		return "Syncing your data..."
	}
}

// SetSize sets the width and height available for rendering.
func (s *Step) SetSize(width, height int) {
	s.width = width
}

// IsBusy returns true while waiting for sync.
func (s *Step) IsBusy() bool {
	return !s.ready
}

// HasError returns false - sync errors are handled by TUI.
func (s *Step) HasError() bool {
	return false
}

// Error returns nil.
func (s *Step) Error() error {
	return nil
}

// Next returns nil when sync is ready (flow complete).
func (s *Step) Next() (step.Step, error) {
	if !s.ready {
		return nil, step.ErrNotReady
	}
	// Flow complete - transition to app
	return nil, nil
}

// Help returns empty key bindings - no actions available.
func (s *Step) Help() help.KeyMap {
	return keymap.Simple{Keys: []key.Binding{}}
}

// Organization returns the organization.
func (s *Step) Organization() api.Organization {
	return s.org
}

// Account returns the account.
func (s *Step) Account() api.Account {
	return s.account
}

// Workspace returns the workspace.
func (s *Step) Workspace() api.Workspace {
	return s.workspace
}

// Close releases any resources held by the step.
func (s *Step) Close() error {
	return nil
}
