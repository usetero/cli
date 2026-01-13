package datadog

import (
	"context"
	"fmt"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/progress"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/complete"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

const (
	pollInterval = 2 * time.Second
)

type tickMsg time.Time

// StatusPoller polls for Datadog account discovery status
type StatusPoller interface {
	GetStatus(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error)
}

type statusFetchedMsg struct {
	status *api.DatadogAccountStatus
	err    error
}

// DiscoveryStep shows unified discovery progress with a single progress bar
type DiscoveryStep struct {
	// Accumulated state from previous steps
	role             string
	orgID            string
	accountID        string
	datadogAccountID *string

	// Services
	statusPoller StatusPoller
	logger       log.Logger

	// Pass-through to next step
	globalBindings []key.Binding

	// UI state
	loading bool
	err     error
	status  *api.DatadogAccountStatus
	spinner spinner.Model
	width   int
}

// NewDiscoveryStep creates a new unified discovery step
func NewDiscoveryStep(
	role string,
	orgID string,
	accountID string,
	datadogAccountID *string,
	statusPoller StatusPoller,
	logger log.Logger,
	globalBindings []key.Binding,
) step.Step {
	if statusPoller == nil {
		panic("statusPoller cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.CurrentTheme().Accent)

	return &DiscoveryStep{
		role:             role,
		orgID:            orgID,
		accountID:        accountID,
		datadogAccountID: datadogAccountID,
		statusPoller:     statusPoller,
		logger:           logger,
		globalBindings:   globalBindings,
		loading:          true,
		spinner:          s,
		width:            80,
	}
}

// Init starts the discovery process
func (s *DiscoveryStep) Init() tea.Cmd {
	return s.startPolling()
}

// startPolling resets state and starts the polling loop with spinner
func (s *DiscoveryStep) startPolling() tea.Cmd {
	s.err = nil
	s.loading = true
	return tea.Batch(
		s.spinner.Tick,
		s.fetchStatus(),
		s.tick(),
	)
}

// tick returns a command that sends a tick message after the poll interval
func (s *DiscoveryStep) tick() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// fetchStatus queries the control plane for discovery status
func (s *DiscoveryStep) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		if s.datadogAccountID == nil {
			return statusFetchedMsg{err: fmt.Errorf("no datadog account specified")}
		}

		s.logger.Debug("fetching discovery status",
			log.String("datadogAccountID", *s.datadogAccountID))

		ctx := context.Background()

		status, err := s.statusPoller.GetStatus(ctx, *s.datadogAccountID)
		if err != nil {
			s.logger.Error("failed to fetch discovery status", "error", err)
			return statusFetchedMsg{err: err}
		}

		if status == nil {
			return statusFetchedMsg{err: fmt.Errorf("no status found")}
		}

		s.logger.Debug("discovery status",
			log.String("status", string(status.Status)),
			log.Int("ready", status.Ready),
			log.Int("total", status.Total))

		return statusFetchedMsg{status: status}
	}
}

// Update handles messages
func (s *DiscoveryStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case statusFetchedMsg:
		wasLoading := s.loading
		s.loading = false

		if msg.err != nil {
			s.err = msg.err
		} else {
			s.status = msg.status
			s.err = nil

			if wasLoading {
				s.logger.Info("discovery started")
			}

			// Log errors (but keep polling)
			if s.status.ErrorMessage != "" {
				s.logger.Warn("discovery has errors",
					log.String("error", s.status.ErrorMessage))
			}

			// Log when discovery is complete
			if s.isComplete() {
				s.logger.Info("discovery completed",
					log.String("status", string(s.status.Status)),
					log.Int("ready", s.status.Ready),
					log.Int("total", s.status.Total))
			}
		}

	case tickMsg:
		// Poll again unless complete or error
		if !s.isComplete() && s.err == nil {
			cmds = append(cmds, s.fetchStatus(), s.tick())
		}

	case tea.KeyMsg:
		// Allow retry on error
		if s.err != nil && msg.String() == "enter" {
			return s, s.startPolling()
		}

	default:
		// Update spinner
		if !s.isComplete() && s.err == nil {
			var cmd tea.Cmd
			s.spinner, cmd = s.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return s, tea.Batch(cmds...)
}

// isComplete returns true if discovery is ready
func (s *DiscoveryStep) isComplete() bool {
	return s.status != nil && s.status.Status == api.DatadogAccountStatusReady
}

// View renders the discovery UI
func (s *DiscoveryStep) View() string {
	theme := styles.CurrentTheme()

	if s.loading {
		return s.renderLoading(theme)
	}

	if s.err != nil {
		return s.renderError(theme)
	}

	return s.renderInProgress(theme)
}

// renderLoading renders the initial loading state
func (s *DiscoveryStep) renderLoading(theme *styles.Theme) string {
	common := styles.Common()

	title := common.Title.Render("Setting up your account...")
	subtitle := common.Subtitle.Render("Discovering services and analyzing log patterns.")
	statusMsg := s.spinner.View() + " " + common.Body.Render("Connecting to Datadog...")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		subtitle,
		"",
		statusMsg,
	)
}

// renderError renders the error state
// Note: The actual error message is displayed by the footer via Error()
func (s *DiscoveryStep) renderError(theme *styles.Theme) string {
	common := styles.Common()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		common.Title.Render("Setting up your account..."),
		"",
		common.Body.Render("Something went wrong."),
		"",
		common.Help.Render("Press enter to retry"),
	)
}

// renderInProgress renders the discovery in progress state
func (s *DiscoveryStep) renderInProgress(theme *styles.Theme) string {
	common := styles.Common()

	title := common.Title.Render("Setting up your account...")
	subtitle := common.Subtitle.Render("Discovering services and analyzing log patterns.")

	// Build status message
	var statusText string
	if s.status == nil {
		statusText = "Starting..."
	} else {
		switch s.status.Status {
		case api.DatadogAccountStatusPending:
			statusText = "Starting discovery..."
		case api.DatadogAccountStatusInProgress:
			statusText = fmt.Sprintf("%d of %d services ready", s.status.Ready, s.status.Total)
		case api.DatadogAccountStatusError:
			statusText = "Discovery encountered an issue"
		default:
			statusText = fmt.Sprintf("%d services discovered", s.status.Total)
		}
	}
	statusMsg := s.spinner.View() + " " + common.Body.Render(statusText)

	// Progress bar
	var progressBar string
	if s.status != nil && s.status.PercentComplete > 0 {
		prog := progress.New(60)
		progressBar = prog.ViewAs(s.status.PercentComplete) // API returns 0-100, ViewAs expects 0-100
	}

	info := common.Help.Render(`
What you'll see next: Waste patterns we found, what's safe to remove,
and one-click actions to improve quality.`)

	elements := []string{
		title,
		"",
		subtitle,
		"",
		"",
		statusMsg,
	}

	if progressBar != "" {
		elements = append(elements, "", progressBar)
	}

	elements = append(elements, "", "", info)

	return lipgloss.JoinVertical(lipgloss.Left, elements...)
}

// SetSize sets the width and height available for rendering
func (s *DiscoveryStep) SetSize(width, height int) {
	s.width = width
}

// IsComplete returns true when discovery has completed
func (s *DiscoveryStep) IsComplete() bool {
	return s.isComplete()
}

// IsBusy returns true while actively discovering
func (s *DiscoveryStep) IsBusy() bool {
	return !s.isComplete() && s.err == nil
}

// HasError returns true if there's an error
func (s *DiscoveryStep) HasError() bool {
	return s.err != nil || (s.status != nil && s.status.ErrorMessage != "")
}

// Error returns the current error
func (s *DiscoveryStep) Error() error {
	if s.err != nil {
		return s.err
	}
	if s.status != nil && s.status.ErrorMessage != "" {
		return fmt.Errorf("%s", s.status.ErrorMessage)
	}
	return nil
}

// Next returns the next step after discovery
func (s *DiscoveryStep) Next() step.Step {
	return complete.NewCompleteStep(s.logger, s.globalBindings)
}

// Help returns the key bindings for this step
func (s *DiscoveryStep) Help() help.KeyMap {
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
	return keymap.Simple{Keys: []key.Binding{}}
}
