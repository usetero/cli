package log_events

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components/progress"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/complete"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
)

const (
	pollInterval = 2 * time.Second
)

type tickMsg time.Time

// LogDiscoveryProgressPoller polls for log event discovery progress
type LogDiscoveryProgressPoller interface {
	GetLogDiscoveryProgress(ctx context.Context, datadogAccountID string) (*api.LogEventDiscoveryProgress, error)
}

// DiscoveryStep shows account-level log event discovery progress with a single progress bar
type DiscoveryStep struct {
	// Accumulated state from previous steps
	role             string
	orgID            string
	accountID        string
	datadogAccountID *string

	// Services
	progressPoller LogDiscoveryProgressPoller

	// Pass-through to next step
	apiClient      api.Client
	logger         log.Logger
	globalBindings []key.Binding

	// UI state
	loading                bool
	err                    error
	discoveryStatus        api.DiscoveryStatus
	discoveryError         string // Last error from discovery job (recoverable)
	percentComplete        float64
	weeklyVolume           int64
	discoveredWeeklyVolume float64
	spinner                spinner.Model
	width                  int
}

// NewDiscoveryStep creates a new log event discovery step
func NewDiscoveryStep(
	role string,
	orgID string,
	accountID string,
	datadogAccountID *string,
	apiClient api.Client,
	logger log.Logger,
	globalBindings []key.Binding,
) step.Step {
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	progressPoller := api.NewDatadogAccountService(apiClient, logger)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.CurrentTheme().Primary)

	return &DiscoveryStep{
		role:             role,
		orgID:            orgID,
		accountID:        accountID,
		datadogAccountID: datadogAccountID,
		progressPoller:   progressPoller,
		apiClient:        apiClient,
		logger:           logger,
		globalBindings:   globalBindings,
		loading:          true,
		discoveryStatus:  api.DiscoveryStatusPending,
		spinner:          s,
		width:            80,
	}
}

// Init starts the discovery process
func (s *DiscoveryStep) Init() tea.Cmd {
	// Start spinner, fetch initial status, and begin polling
	return tea.Batch(
		s.spinner.Tick,
		s.fetchProgress(),
		s.tick(),
	)
}

// tick returns a command that sends a tick message after the poll interval
func (s *DiscoveryStep) tick() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type progressFetchedMsg struct {
	status                 api.DiscoveryStatus
	percentComplete        float64
	weeklyVolume           int64
	discoveredWeeklyVolume float64
	discoveryError         string // Last error from discovery job
	err                    error  // API/system error
}

// fetchProgress queries the control plane for log discovery progress
func (s *DiscoveryStep) fetchProgress() tea.Cmd {
	return func() tea.Msg {
		if s.datadogAccountID == nil {
			return progressFetchedMsg{err: fmt.Errorf("no datadog account specified")}
		}

		s.logger.Debug("fetching log event discovery progress",
			log.String("datadogAccountID", *s.datadogAccountID))

		ctx := context.Background()

		progress, err := s.progressPoller.GetLogDiscoveryProgress(ctx, *s.datadogAccountID)
		if err != nil {
			s.logger.Error("failed to fetch discovery progress", "error", err)
			return progressFetchedMsg{err: err}
		}

		if progress == nil {
			return progressFetchedMsg{err: fmt.Errorf("no discovery progress found")}
		}

		percent := 0.0
		if progress.PercentComplete != nil {
			percent = *progress.PercentComplete
		}

		s.logger.Debug("discovery progress",
			log.String("status", string(progress.Status)),
			log.Float64("percentComplete", percent))

		return progressFetchedMsg{
			status:                 progress.Status,
			percentComplete:        percent,
			weeklyVolume:           progress.WeeklyVolume,
			discoveredWeeklyVolume: progress.DiscoveredWeeklyVolume,
			discoveryError:         progress.LastError,
		}
	}
}

// Update handles messages
func (s *DiscoveryStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case progressFetchedMsg:
		wasLoading := s.loading
		s.loading = false

		if msg.err != nil {
			s.err = msg.err
		} else {
			s.discoveryStatus = msg.status
			s.percentComplete = msg.percentComplete
			s.weeklyVolume = msg.weeklyVolume
			s.discoveredWeeklyVolume = msg.discoveredWeeklyVolume
			s.discoveryError = msg.discoveryError

			if wasLoading {
				s.logger.Info("log event discovery started")
			}

			// Log discovery errors (but keep polling)
			if s.discoveryError != "" {
				s.logger.Warn("log event discovery has errors",
					log.String("error", s.discoveryError))
			}

			// Log and fast-forward when discovery is complete
			if s.isComplete() {
				s.logger.Info("log event discovery completed, transitioning to app",
					log.String("status", string(s.discoveryStatus)))
				// Flow will detect IsComplete() and auto-advance
			}
		}

	case tickMsg:
		// Poll again unless complete or error
		if !s.isComplete() && s.err == nil {
			cmds = append(cmds, s.fetchProgress(), s.tick())
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

// isComplete returns true if discovery is ready (successfully completed)
func (s *DiscoveryStep) isComplete() bool {
	return s.discoveryStatus == api.DiscoveryStatusReady
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

	// No completion screen - auto-transitions
	return s.renderInProgress(theme)
}

// renderLoading renders the initial loading state
func (s *DiscoveryStep) renderLoading(theme *styles.Theme) string {
	common := styles.Common()

	title := common.Body.Bold(true).Render("Understanding your log patterns and identifying waste...")
	subtitle := common.Subtitle.Render("This may take a few minutes depending on volume.")
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
func (s *DiscoveryStep) renderError(theme *styles.Theme) string {
	common := styles.Common()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		common.Title.Render("Analyzing your top services for waste..."),
		"",
		common.Error.Render("Error: "+s.err.Error()),
	)
}

// renderInProgress renders the discovery in progress state
func (s *DiscoveryStep) renderInProgress(theme *styles.Theme) string {
	common := styles.Common()

	title := common.Body.Bold(true).Render("Understanding your log patterns and identifying waste...")
	subtitle := common.Subtitle.Render("This may take a few minutes depending on volume.")

	// Create progress bar with gradient - match content width
	prog := progress.New(60)
	progressBarWithPercent := prog.ViewAs(s.percentComplete)

	// Format log counts with spinner
	logCounts := s.formatLogCounts()
	statusMsg := s.spinner.View() + " " + common.Body.Render(logCounts)

	info := common.Help.Render(`
What you'll see next: Waste patterns we found, what's safe to remove,
and one-click actions to improve quality.`)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		subtitle,
		"",
		"",
		statusMsg,
		"",
		progressBarWithPercent,
		"",
		"",
		info,
	)
}

// formatLogCounts formats the log counts in a human-readable way
func (s *DiscoveryStep) formatLogCounts() string {
	if s.weeklyVolume == 0 {
		return "Starting analysis..."
	}

	discovered := int64(s.discoveredWeeklyVolume)
	total := s.weeklyVolume

	return fmt.Sprintf("Analyzed %s / %s logs", formatVolume(discovered), formatVolume(total))
}

// formatVolume formats a volume number in a human-readable way (e.g., 4.6M, 7.6M)
func formatVolume(volume int64) string {
	if volume < 1000 {
		return fmt.Sprintf("%d", volume)
	} else if volume < 1000000 {
		return fmt.Sprintf("%.1fK", float64(volume)/1000)
	} else if volume < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(volume)/1000000)
	}
	return fmt.Sprintf("%.1fB", float64(volume)/1000000000)
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

// HasError returns true if there's a discovery error or API error
func (s *DiscoveryStep) HasError() bool {
	return s.err != nil || s.discoveryError != ""
}

// Error returns the current error (discovery or API error)
func (s *DiscoveryStep) Error() error {
	if s.err != nil {
		return s.err
	}
	if s.discoveryError != "" {
		return fmt.Errorf("%s", s.discoveryError)
	}
	return nil
}

// Next returns the next step after discovery
func (s *DiscoveryStep) Next() step.Step {
	// Show completion message
	return complete.NewCompleteStep(s.logger, s.globalBindings)
}

// Help returns the key bindings for this step
func (s *DiscoveryStep) Help() help.KeyMap {
	// No bindings during discovery - auto-transitions when complete
	return keymap.Simple{Keys: []key.Binding{}}
}
