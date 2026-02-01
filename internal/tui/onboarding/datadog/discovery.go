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

type statusFetchedMsg struct {
	status *api.DatadogAccountStatus
	err    error
}

// DiscoveryStep shows discovery progress focused on what users care about
type DiscoveryStep struct {
	// Context for API calls
	ctx context.Context

	// Theme for styling
	theme *styles.Theme

	// Accumulated state from previous steps
	role             string
	org              api.Organization
	account          api.Account
	datadogAccountID *string

	// Services
	datadogAccounts api.DatadogAccounts
	logger          log.Logger
	globalBindings  []key.Binding

	// UI state
	loading   bool
	err       error
	status    *api.DatadogAccountStatus
	spinner   spinner.Model
	width     int
	startTime time.Time
}

// NewDiscoveryStep creates a new discovery step
func NewDiscoveryStep(
	ctx context.Context,
	theme *styles.Theme,
	role string,
	org api.Organization,
	account api.Account,
	datadogAccountID *string,
	datadogAccounts api.DatadogAccounts,
	logger log.Logger,
	globalBindings []key.Binding,
) step.Step {
	if datadogAccounts == nil {
		panic("datadogAccounts cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	return &DiscoveryStep{
		ctx:              ctx,
		theme:            theme,
		role:             role,
		org:              org,
		account:          account,
		datadogAccountID: datadogAccountID,
		datadogAccounts:  datadogAccounts,
		logger:           logger,
		globalBindings:   globalBindings,
		loading:          true,
		spinner:          s,
		width:            80,
		startTime:        time.Now(),
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
	s.startTime = time.Now()
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

		status, err := s.datadogAccounts.GetStatus(s.ctx, *s.datadogAccountID)
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

// isComplete returns true when the account has enough data to proceed.
// The control plane determines this via ReadyForUse - we don't encode
// business logic about thresholds in the CLI.
func (s *DiscoveryStep) isComplete() bool {
	if s.status == nil {
		return false
	}
	return s.status.ReadyForUse
}

// View renders the discovery UI
func (s *DiscoveryStep) View() string {
	themeStyles := s.theme.Styles
	colors := s.theme.Colors

	if s.loading || s.status == nil {
		return s.renderLoading(themeStyles)
	}

	if s.err != nil {
		return s.renderError(themeStyles)
	}

	// Title
	title := themeStyles.Title.Render("Analyzing your Datadog account")

	// Main status line - what's actually happening
	statusLine := s.spinner.View() + " " + s.getStatusText(themeStyles)

	// Progress section
	var progressSection string
	if s.status.ServiceCount > 0 {
		progressSection = s.renderProgress(themeStyles, colors)
	}

	// Issues section - surface problems prominently
	issuesSection := s.renderIssues(themeStyles, colors)

	// Help text
	helpText := s.getHelpText(themeStyles)

	// Build the view
	parts := []string{title, "", statusLine}

	if progressSection != "" {
		parts = append(parts, "", progressSection)
	}

	if issuesSection != "" {
		parts = append(parts, "", issuesSection)
	}

	if helpText != "" {
		parts = append(parts, "", helpText)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// getStatusText returns the main status message
func (s *DiscoveryStep) getStatusText(themeStyles *styles.Styles) string {
	st := s.status

	// No services yet
	if st.ServiceCount == 0 {
		return themeStyles.Body.Render("Looking for services...")
	}

	// All inactive - no logs (user needs to send data)
	if st.Status == api.DatadogAccountStatusInactive {
		return themeStyles.Body.Render(fmt.Sprintf("Found %d services with no recent log data", st.ServiceCount))
	}

	// Stale - our system hasn't run (our problem)
	if st.Status == api.DatadogAccountStatusStale {
		return themeStyles.Body.Render(fmt.Sprintf("Found %d services, but our data is out of date", st.ServiceCount))
	}

	// Still discovering - waiting on our analysis pipeline
	if st.DiscoveringServices > 0 && st.AnalyzingServices == 0 && st.ReadyServices == 0 {
		return themeStyles.Body.Render(fmt.Sprintf("Found %d services, waiting for analysis...", st.ServiceCount))
	}

	// Actively analyzing
	if st.AnalyzingServices > 0 || st.ReadyServices > 0 {
		if st.ServiceLogVolume > 0 {
			return themeStyles.Body.Render(fmt.Sprintf("Analyzing %s logs from %d services...",
				formatVolume(st.ServiceLogVolume), st.ActiveServices))
		}
		return themeStyles.Body.Render(fmt.Sprintf("Analyzing %d services...", st.ActiveServices))
	}

	// Fallback
	return themeStyles.Body.Render(fmt.Sprintf("Processing %d services...", st.ServiceCount))
}

// renderProgress renders the progress bar and log event counts
func (s *DiscoveryStep) renderProgress(themeStyles *styles.Styles, colors *styles.Colors) string {
	st := s.status

	// Progress is based on saved count toward the ready_for_use threshold (50)
	const readyThreshold = 50
	pct := float64(st.SavedCount) / float64(readyThreshold)
	if pct > 1.0 {
		pct = 1.0
	}

	// Progress bar
	prog := progress.New(s.theme, 50)
	progressBar := prog.ViewAs(pct)

	// Show saved count progress
	readyStyle := lipgloss.NewStyle().Foreground(colors.Success.Fg)
	mutedStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)

	countText := fmt.Sprintf("%s / %s",
		readyStyle.Render(fmt.Sprintf("%d", st.SavedCount)),
		mutedStyle.Render(fmt.Sprintf("%d log events", readyThreshold)))

	return lipgloss.JoinVertical(lipgloss.Left, progressBar, "", countText)
}

// renderIssues surfaces any problems the user should know about
func (s *DiscoveryStep) renderIssues(themeStyles *styles.Styles, colors *styles.Colors) string {
	st := s.status
	var issues []string

	warningStyle := lipgloss.NewStyle().Foreground(colors.Warning.Fg)
	errorStyle := lipgloss.NewStyle().Foreground(colors.Error.Fg)

	// Stale data - this is OUR problem, not theirs
	if st.StaleServices > 0 || st.Status == api.DatadogAccountStatusStale {
		issues = append(issues, warningStyle.Render("! Our analysis is more than 48 hours old"))
	}

	// Broken services
	if st.BrokenServices > 0 {
		issues = append(issues, errorStyle.Render(
			fmt.Sprintf("! %d services have errors", st.BrokenServices)))
	}

	// Inactive - no log data (user's issue, but only show if not the main status)
	if st.InactiveServices > 0 && st.Status != api.DatadogAccountStatusInactive {
		issues = append(issues, warningStyle.Render(
			fmt.Sprintf("! %d services have no recent logs", st.InactiveServices)))
	}

	if len(issues) == 0 {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Left, issues...)
}

// getHelpText returns contextual help based on status
func (s *DiscoveryStep) getHelpText(themeStyles *styles.Styles) string {
	st := s.status

	// All inactive - user needs to send data
	if st.Status == api.DatadogAccountStatusInactive {
		return themeStyles.Help.Render("Send logs to Datadog to continue. We look at the last 7 days of data.")
	}

	// Stale - our problem
	if st.Status == api.DatadogAccountStatusStale {
		return themeStyles.Help.Render("This is on our end. We're working on it - try again soon.")
	}

	// Still waiting on our pipeline after finding services
	if st.DiscoveringServices > 0 && st.AnalyzingServices == 0 && st.ReadyServices == 0 {
		elapsed := time.Since(s.startTime)
		if elapsed > 30*time.Second {
			return themeStyles.Help.Render("Taking longer than expected. Our system may be catching up.")
		}
	}

	// Normal progress
	if st.AnalyzingServices > 0 {
		return themeStyles.Help.Render("Initial analysis takes a few minutes.")
	}

	return ""
}

// renderLoading renders the initial loading state
func (s *DiscoveryStep) renderLoading(themeStyles *styles.Styles) string {
	title := themeStyles.Title.Render("Analyzing your Datadog account")
	statusMsg := s.spinner.View() + " " + themeStyles.Body.Render("Connecting...")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		statusMsg,
	)
}

// renderError renders the error state
func (s *DiscoveryStep) renderError(themeStyles *styles.Styles) string {
	title := themeStyles.Title.Render("Analyzing your Datadog account")
	statusMsg := themeStyles.Body.Render("Something went wrong.")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		statusMsg,
		"",
		themeStyles.Help.Render("Press enter to retry"),
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

// HasError returns true if there's an unrecoverable error.
// Broken services are shown as warnings, not errors - users can still proceed.
func (s *DiscoveryStep) HasError() bool {
	if s.err != nil {
		return true
	}
	if s.status == nil {
		return false
	}
	// All services disabled - nothing to analyze
	return s.status.Status == api.DatadogAccountStatusDisabled
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
		return fmt.Errorf("all services are disabled")
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
	return complete.NewCompleteStep(s.ctx, s.theme, s.org, s.account, s.logger, s.globalBindings), nil
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
