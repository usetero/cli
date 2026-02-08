package datadog

import (
	"context"
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/progress"
)

const discoveryPollInterval = 500 * time.Millisecond

// statusMsg is sent on each poll.
type statusMsg struct {
	status *api.DatadogAccountStatus
	err    error
}

// DiscoveryModel polls for Datadog discovery status.
type DiscoveryModel struct {
	ctx              context.Context
	theme            styles.Theme
	services         api.APIServices
	scope            log.Scope
	datadogAccountID domain.DatadogAccountID

	spinner  spinner.Model
	progress *progress.Model
	status   *api.DatadogAccountStatus
	err      error
	width    int
	height   int
}

// NewDiscovery creates a new discovery step.
func NewDiscovery(
	ctx context.Context,
	theme styles.Theme,
	datadogAccountID domain.DatadogAccountID,
	services api.APIServices,
	scope log.Scope,
) *DiscoveryModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if datadogAccountID == "" {
		panic("datadogAccountID is empty")
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Accent)

	return &DiscoveryModel{
		ctx:              ctx,
		theme:            theme,
		services:         services,
		scope:            scope,
		datadogAccountID: datadogAccountID,
		spinner:          sp,
		progress:         progress.New(theme, 50),
	}
}

// Init starts polling for status.
func (m *DiscoveryModel) Init() tea.Cmd {
	m.scope.Info("starting datadog discovery", "datadogAccountID", m.datadogAccountID)
	return tea.Batch(m.spinner.Tick, m.pollStatus())
}

func (m *DiscoveryModel) pollStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.services.DatadogAccounts.GetStatus(m.ctx, m.datadogAccountID.String())
		return statusMsg{status: status, err: err}
	}
}

func (m *DiscoveryModel) schedulePoll() tea.Cmd {
	return tea.Tick(discoveryPollInterval, func(time.Time) tea.Msg {
		status, err := m.services.DatadogAccounts.GetStatus(m.ctx, m.datadogAccountID.String())
		return statusMsg{status: status, err: err}
	})
}

// Update handles messages.
func (m *DiscoveryModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case statusMsg:
		if msg.err != nil {
			m.scope.Error("discovery status check failed", "error", msg.err)
			m.err = msg.err
			return appmsg.ErrorCmd("Failed to check discovery status", msg.err, false)
		}
		m.status = msg.status

		if msg.status.ReadyForUse {
			m.scope.Info("datadog discovery complete", "services", msg.status.ServiceCount)
			return func() tea.Msg { return msgs.DatadogDiscoveryComplete{} }
		}
		return m.schedulePoll()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd

	case tea.KeyPressMsg:
		if msg.String() == "r" && m.err != nil {
			m.scope.Debug("retrying discovery")
			m.err = nil
			return m.Init()
		}
	}

	return m.progress.Update(msg)
}

// View renders the discovery UI.
func (m *DiscoveryModel) View() string {
	s := m.theme.Styles

	title := s.Title.Render("Discovering Datadog services")

	if m.err != nil {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			s.Error.Render(fmt.Sprintf("Discovery failed: %v", m.err)),
			"",
			s.Help.Render("Press 'r' to retry"),
		)
	}

	if m.status == nil {
		statusLine := m.spinner.View() + " " + s.Body.Render("Connecting...")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)
	}

	if m.status.ReadyForUse {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			s.Success.Render("Discovery complete!"),
		)
	}

	statusText := m.statusText()
	statusLine := m.spinner.View() + " " + s.Body.Render(statusText)
	progressBar := m.progress.ViewAs(m.status.PercentComplete)

	parts := []string{title, "", statusLine, "", progressBar}

	if m.status.ServiceCount > 0 {
		countText := fmt.Sprintf("%d / %d services analyzed", m.status.ReadyServices, m.status.ActiveServices)
		parts = append(parts, "", s.Help.Render(countText))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *DiscoveryModel) statusText() string {
	switch m.status.Status {
	case api.DatadogAccountStatusDiscovering:
		return "Discovering services..."
	case api.DatadogAccountStatusAnalyzing:
		return "Analyzing log patterns..."
	case api.DatadogAccountStatusReady:
		return "Ready!"
	default:
		return "Processing..."
	}
}

// SetSize updates dimensions.
func (m *DiscoveryModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.progress.SetWidth(min(width, 50))
}

// ShortHelp returns the key bindings for the short help view.
func (m *DiscoveryModel) ShortHelp() []key.Binding {
	if m.err != nil {
		return []key.Binding{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		}
	}
	return nil
}
