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

// formatVolume formats a volume count into a human-readable string (e.g., "2.4M", "150K")
func formatVolume(volume int) string {
	switch {
	case volume >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(volume)/1_000_000_000)
	case volume >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(volume)/1_000_000)
	case volume >= 1_000:
		return fmt.Sprintf("%.1fK", float64(volume)/1_000)
	default:
		return fmt.Sprintf("%d", volume)
	}
}

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
	org              api.Organization
	account          api.Account
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
	org api.Organization,
	account api.Account,
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
		org:              org,
		account:          account,
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
			log.Int("ready", status.ReadyServices),
			log.Int("total", status.ServiceCount))

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
			if s.status.BrokenServices > 0 {
				s.logger.Warn("discovery has broken services",
					log.Int("broken", s.status.BrokenServices))
			}

			// Log when discovery is complete
			if s.isComplete() {
				s.logger.Info("discovery completed",
					log.String("status", string(s.status.Status)),
					log.Int("ready", s.status.ReadyServices),
					log.Int("total", s.status.ServiceCount))
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

// isAllInactive returns true if all services are inactive (zero volume)
func (s *DiscoveryStep) isAllInactive() bool {
	return s.status != nil && s.status.Status == api.DatadogAccountStatusInactive
}

// View renders the discovery UI
func (s *DiscoveryStep) View() string {
	common := styles.Common()

	if s.loading {
		return s.renderLoading(common)
	}

	if s.err != nil {
		return s.renderError(common)
	}

	if s.status == nil {
		return s.renderLoading(common)
	}

	// Render based on status
	switch s.status.Status {
	case api.DatadogAccountStatusDiscovering:
		return s.renderDiscovering(common)
	case api.DatadogAccountStatusAnalyzing:
		return s.renderInProgress(common)
	case api.DatadogAccountStatusInactive:
		return s.renderInactive(common)
	case api.DatadogAccountStatusDisabled:
		return s.renderDisabled(common)
	case api.DatadogAccountStatusBroken:
		return s.renderError(common)
	default:
		return s.renderInProgress(common)
	}
}

// renderLoading renders the initial loading state
func (s *DiscoveryStep) renderLoading(common *styles.CommonStyles) string {
	title := common.Title.Render("Setting up your account...")
	statusMsg := s.spinner.View() + " " + common.Body.Render("Connecting to Datadog...")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		statusMsg,
	)
}

// renderDiscovering renders the discovering state (no volume data yet)
func (s *DiscoveryStep) renderDiscovering(common *styles.CommonStyles) string {
	title := common.Title.Render("Setting up your account...")
	statusMsg := s.spinner.View() + " " + common.Body.Render("Discovering services...")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		statusMsg,
	)
}

// renderInProgress renders the discovery in progress state
func (s *DiscoveryStep) renderInProgress(common *styles.CommonStyles) string {
	title := common.Title.Render("Setting up your account...")

	// Build status message with volume if available
	var statusText string
	if s.status.ServiceLogVolume > 0 && s.status.DiscoveredLogVolume > 0 {
		statusText = fmt.Sprintf("Analyzed %s of %s logs across %d services...",
			formatVolume(s.status.DiscoveredLogVolume),
			formatVolume(s.status.ServiceLogVolume),
			s.status.ServiceCount)
	} else if s.status.ServiceLogVolume > 0 {
		statusText = fmt.Sprintf("Analyzing %s logs across %d services...", formatVolume(s.status.ServiceLogVolume), s.status.ServiceCount)
	} else {
		statusText = fmt.Sprintf("Analyzing %d services...", s.status.ServiceCount)
	}
	statusMsg := s.spinner.View() + " " + common.Body.Render(statusText)

	// Progress bar
	prog := progress.New(60)
	progressBar := prog.ViewAs(s.status.PercentComplete)

	// Subtle hint about timing
	hint := common.Help.Render("This initial analysis usually takes a few minutes.")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		statusMsg,
		"",
		progressBar,
		"",
		hint,
	)
}

// renderInactive renders the state when all services have zero log volume
func (s *DiscoveryStep) renderInactive(common *styles.CommonStyles) string {
	theme := styles.CurrentTheme()
	title := common.Title.Render("Setting up your account...")

	statusText := fmt.Sprintf("Found %d services, but no recent log data", s.status.InactiveServices)
	statusMsg := s.spinner.View() + " " + common.Body.Render(statusText)

	// Yellow substatus for user action needed
	warningStyle := lipgloss.NewStyle().Foreground(theme.Warning.Fg)
	substatus := warningStyle.Render("⚠ Send logs to Datadog to continue. Looking for data from the last 7 days.")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		statusMsg,
		"",
		substatus,
	)
}

// renderDisabled renders the error state when all services are disabled
func (s *DiscoveryStep) renderDisabled(common *styles.CommonStyles) string {
	title := common.Title.Render("Setting up your account...")
	statusMsg := common.Body.Render("All services are disabled")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		statusMsg,
		"",
		common.Help.Render("Press enter to retry"),
	)
}

// renderError renders the error state
// Note: The actual error message is displayed by the footer via Error()
func (s *DiscoveryStep) renderError(common *styles.CommonStyles) string {
	title := common.Title.Render("Setting up your account...")
	statusMsg := common.Body.Render("Something went wrong.")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		statusMsg,
		"",
		common.Help.Render("Press enter to retry"),
	)
}

// SetSize sets the width and height available for rendering
func (s *DiscoveryStep) SetSize(width, height int) {
	s.width = width
}

// IsBusy returns true while actively discovering
func (s *DiscoveryStep) IsBusy() bool {
	return !s.isComplete() && s.err == nil
}

// HasError returns true if there's an error (including DISABLED and BROKEN statuses)
func (s *DiscoveryStep) HasError() bool {
	if s.err != nil {
		return true
	}
	if s.status == nil {
		return false
	}
	return s.status.Status == api.DatadogAccountStatusDisabled ||
		s.status.Status == api.DatadogAccountStatusBroken
}

// Error returns the current error
func (s *DiscoveryStep) Error() error {
	if s.err != nil {
		return s.err
	}
	if s.status == nil {
		return nil
	}
	if s.status.Status == api.DatadogAccountStatusDisabled {
		return fmt.Errorf("All services are disabled. Enable services to continue.")
	}
	if s.status.Status == api.DatadogAccountStatusBroken {
		return fmt.Errorf("Discovery failed")
	}
	return nil
}

// Next returns the next step after discovery
func (s *DiscoveryStep) Next() (step.Step, error) {
	if s.err != nil {
		return nil, s.err
	}
	if !s.isComplete() {
		return nil, step.ErrNotReady
	}
	return complete.NewCompleteStep(s.org, s.account, s.logger, s.globalBindings), nil
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
