package datadogdiscovery

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
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/progress"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	workspaceselect "github.com/usetero/cli/internal/tui/onboarding/workspace/select"
)

const pollInterval = 2 * time.Second

type tickMsg time.Time

type statusFetchedMsg struct {
	status *api.DatadogAccountStatus
	err    error
}

// formatVolume formats a volume count into a human-readable string.
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

// Model shows discovery progress.
type Model struct {
	ctx   context.Context
	theme *styles.Theme

	role             string
	org              domain.Organization
	account          domain.Account
	datadogAccountID *string

	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger

	loading   bool
	err       error
	status    *api.DatadogAccountStatus
	spinner   spinner.Model
	startTime time.Time
	width     int
	height    int
}

// New creates a new discovery model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	role string,
	org domain.Organization,
	account domain.Account,
	datadogAccountID *string,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	return Model{
		ctx:              ctx,
		theme:            theme,
		role:             role,
		org:              org,
		account:          account,
		datadogAccountID: datadogAccountID,
		services:         services,
		prefs:            prefs,
		logger:           logger,
		loading:          true,
		spinner:          sp,
		startTime:        time.Now(),
		width:            80,
	}
}

// Init starts the discovery process.
func (m Model) Init() tea.Cmd {
	return m.startPolling()
}

// startPolling resets state and starts the polling loop.
func (m Model) startPolling() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchStatus(),
		m.tick(),
	)
}

// tick returns a command that sends a tick message after the poll interval.
func (m Model) tick() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// fetchStatus queries the control plane for discovery status.
func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		if m.datadogAccountID == nil {
			return statusFetchedMsg{err: fmt.Errorf("no datadog account specified")}
		}

		m.logger.Debug("fetching discovery status", "datadogAccountID", *m.datadogAccountID)

		status, err := m.services.DatadogAccounts.GetStatus(m.ctx, *m.datadogAccountID)
		if err != nil {
			m.logger.Error("failed to fetch discovery status", "error", err)
			return statusFetchedMsg{err: err}
		}

		if status == nil {
			return statusFetchedMsg{err: fmt.Errorf("no status found")}
		}

		m.logger.Debug("discovery status", "status", string(status.Status), "ready", status.ReadyServices, "total", status.ServiceCount)
		return statusFetchedMsg{status: status}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case statusFetchedMsg:
		wasLoading := m.loading
		m.loading = false

		if msg.err != nil {
			m.err = msg.err
		} else {
			m.status = msg.status
			m.err = nil

			if wasLoading {
				m.logger.Info("discovery started")
			}

			if m.status.BrokenServices > 0 {
				m.logger.Warn("discovery has broken services", "broken", m.status.BrokenServices)
			}

			if m.isComplete() {
				m.logger.Info("discovery completed", "status", string(m.status.Status), "ready", m.status.ReadyServices, "total", m.status.ServiceCount)
			}
		}

	case tickMsg:
		if !m.isComplete() && m.err == nil {
			cmds = append(cmds, m.fetchStatus(), m.tick())
		}

	case tea.KeyPressMsg:
		if m.err != nil && msg.String() == "enter" {
			m.err = nil
			m.loading = true
			m.startTime = time.Now()
			return m, m.startPolling()
		}

	default:
		if !m.isComplete() && m.err == nil {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// isComplete returns true when the account has enough data to proceed.
func (m Model) isComplete() bool {
	if m.status == nil {
		return false
	}
	return m.status.ReadyForUse
}

// View renders the discovery UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles
	colors := m.theme.Colors

	if m.loading || m.status == nil {
		return m.renderLoading(themeStyles)
	}

	if m.err != nil {
		return m.renderError(themeStyles)
	}

	title := themeStyles.Title.Render("Analyzing your Datadog account")
	statusLine := m.spinner.View() + " " + m.getStatusText(themeStyles)

	var progressSection string
	if m.status.ServiceCount > 0 {
		progressSection = m.renderProgress(themeStyles, colors)
	}

	issuesSection := m.renderIssues(colors)
	helpText := m.getHelpText(themeStyles)

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

// getStatusText returns the main status message.
func (m Model) getStatusText(themeStyles *styles.Styles) string {
	st := m.status

	if st.ServiceCount == 0 {
		return themeStyles.Body.Render("Looking for services...")
	}

	if st.Status == api.DatadogAccountStatusInactive {
		return themeStyles.Body.Render(fmt.Sprintf("Found %d services with no recent log data", st.ServiceCount))
	}

	if st.Status == api.DatadogAccountStatusStale {
		return themeStyles.Body.Render(fmt.Sprintf("Found %d services, but our data is out of date", st.ServiceCount))
	}

	if st.DiscoveringServices > 0 && st.AnalyzingServices == 0 && st.ReadyServices == 0 {
		return themeStyles.Body.Render(fmt.Sprintf("Found %d services, waiting for analysis...", st.ServiceCount))
	}

	if st.AnalyzingServices > 0 || st.ReadyServices > 0 {
		if st.ServiceLogVolume > 0 {
			return themeStyles.Body.Render(fmt.Sprintf("Analyzing %s logs from %d services...", formatVolume(st.ServiceLogVolume), st.ActiveServices))
		}
		return themeStyles.Body.Render(fmt.Sprintf("Analyzing %d services...", st.ActiveServices))
	}

	return themeStyles.Body.Render(fmt.Sprintf("Processing %d services...", st.ServiceCount))
}

// renderProgress renders the progress bar.
func (m Model) renderProgress(themeStyles *styles.Styles, colors *styles.Colors) string {
	st := m.status
	const readyThreshold = 50

	pct := float64(st.SavedCount) / float64(readyThreshold)
	if pct > 1.0 {
		pct = 1.0
	}

	prog := progress.New(m.theme, 50)
	progressBar := prog.ViewAs(pct)

	readyStyle := lipgloss.NewStyle().Foreground(colors.Success.Fg)
	mutedStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	countText := fmt.Sprintf("%s / %s", readyStyle.Render(fmt.Sprintf("%d", st.SavedCount)), mutedStyle.Render(fmt.Sprintf("%d log events", readyThreshold)))

	return lipgloss.JoinVertical(lipgloss.Left, progressBar, "", countText)
}

// renderIssues surfaces any problems.
func (m Model) renderIssues(colors *styles.Colors) string {
	st := m.status
	var issues []string

	warningStyle := lipgloss.NewStyle().Foreground(colors.Warning.Fg)
	errorStyle := lipgloss.NewStyle().Foreground(colors.Error.Fg)

	if st.StaleServices > 0 || st.Status == api.DatadogAccountStatusStale {
		issues = append(issues, warningStyle.Render("! Our analysis is more than 48 hours old"))
	}

	if st.BrokenServices > 0 {
		issues = append(issues, errorStyle.Render(fmt.Sprintf("! %d services have errors", st.BrokenServices)))
	}

	if st.InactiveServices > 0 && st.Status != api.DatadogAccountStatusInactive {
		issues = append(issues, warningStyle.Render(fmt.Sprintf("! %d services have no recent logs", st.InactiveServices)))
	}

	if len(issues) == 0 {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Left, issues...)
}

// getHelpText returns contextual help based on status.
func (m Model) getHelpText(themeStyles *styles.Styles) string {
	st := m.status

	if st.Status == api.DatadogAccountStatusInactive {
		return themeStyles.Help.Render("Send logs to Datadog to continue. We look at the last 7 days of data.")
	}

	if st.Status == api.DatadogAccountStatusStale {
		return themeStyles.Help.Render("This is on our end. We're working on it - try again soon.")
	}

	if st.DiscoveringServices > 0 && st.AnalyzingServices == 0 && st.ReadyServices == 0 {
		elapsed := time.Since(m.startTime)
		if elapsed > 30*time.Second {
			return themeStyles.Help.Render("Taking longer than expected. Our system may be catching up.")
		}
	}

	if st.AnalyzingServices > 0 {
		return themeStyles.Help.Render("Initial analysis takes a few minutes.")
	}

	return ""
}

// renderLoading renders the initial loading state.
func (m Model) renderLoading(themeStyles *styles.Styles) string {
	title := themeStyles.Title.Render("Analyzing your Datadog account")
	statusMsg := m.spinner.View() + " " + themeStyles.Body.Render("Connecting...")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", statusMsg)
}

// renderError renders the error state.
func (m Model) renderError(themeStyles *styles.Styles) string {
	title := themeStyles.Title.Render("Analyzing your Datadog account")
	statusMsg := themeStyles.Body.Render("Something went wrong.")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", statusMsg, "", themeStyles.Help.Render("Press enter to retry"))
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	return m
}

// IsBusy returns true while actively discovering.
func (m Model) IsBusy() bool {
	return !m.isComplete() && m.err == nil
}

// HasError returns true if there's an unrecoverable error.
func (m Model) HasError() bool {
	if m.err != nil {
		return true
	}
	if m.status == nil {
		return false
	}
	return m.status.Status == api.DatadogAccountStatusDisabled
}

// Error returns the current error.
func (m Model) Error() error {
	if m.err != nil {
		return m.err
	}
	if m.status == nil {
		return nil
	}
	if m.status.Status == api.DatadogAccountStatusDisabled {
		return fmt.Errorf("all services are disabled")
	}
	return nil
}

// Help returns the key bindings for this step.
func (m Model) Help() help.KeyMap {
	if m.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "retry")),
			},
		}
	}
	return keymap.Simple{Keys: []key.Binding{}}
}

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.isComplete() {
		return nil, step.ErrNotReady
	}

	return workspaceselect.New(
		m.ctx,
		m.theme,
		m.org,
		m.account,
		m.services,
		m.prefs,
		m.logger,
	), nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
