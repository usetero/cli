package datadogdiscovery

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/interfaces/tui/components/progressbar"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model renders the Datadog discovery wait step.
type Model struct {
	theme    theme.Theme
	progress *progressbar.Model
	status   *integrations.DatadogStatus
	width    int
}

var _ core.Model = (*Model)(nil)
var _ core.InputProvider = (*Model)(nil)

func New(appTheme theme.Theme) *Model {
	return &Model{
		theme:    appTheme,
		progress: progressbar.New(appTheme, 32),
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

func (m *Model) View() tea.View {
	parts := []string{
		m.theme.Text.Section.Render("Discovering Datadog events"),
		"",
		m.theme.Text.Body.Render(m.statusLine()),
	}

	if pct, ok := m.percent(); ok {
		parts = append(parts, "", m.progress.ViewAs(pct))
	}
	if detail := strings.TrimSpace(m.detailLine()); detail != "" {
		parts = append(parts, "", m.theme.Text.Subtle.Render(detail))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return tea.NewView(present.Panel(m.theme.OnSurface(), m.width, content))
}

func (m *Model) SetSize(width, _ int) {
	if width < 1 {
		width = 1
	}
	m.width = width
	m.progress.SetWidth(min(48, present.PanelInnerWidth(width)))
}

func (m *Model) Input() *core.Input {
	return &core.Input{
		Label: "We're discovering your Datadog events. This can take a few minutes.",
	}
}

func (m *Model) SetStatus(status *integrations.DatadogStatus) {
	m.status = status
}

func (m *Model) statusLine() string {
	if m.status == nil {
		return "Connecting to Datadog and preparing discovery."
	}
	if m.status.ReadyForUse {
		return "Discovery is complete."
	}
	if m.status.EventCount > 0 {
		return "Analyzing Datadog events and building your account."
	}
	return "Discovery is running."
}

func (m *Model) detailLine() string {
	if m.status == nil {
		return ""
	}
	if m.status.EventCount > 0 {
		return fmt.Sprintf("%d / %d events analyzed", m.status.AnalyzedCount, m.status.EventCount)
	}
	if m.status.ActiveServices > 0 {
		return fmt.Sprintf("%d active services found", m.status.ActiveServices)
	}
	return ""
}

func (m *Model) percent() (float64, bool) {
	if m.status == nil || m.status.EventCount <= 0 {
		return 0, false
	}
	return float64(m.status.AnalyzedCount) / float64(m.status.EventCount) * 100, true
}
