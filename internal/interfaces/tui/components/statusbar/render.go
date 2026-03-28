package statusbar

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
)

// SetSize satisfies the shared TUI model contract.
func (m *Model) SetSize(width, _ int) {
	if width < 0 {
		width = 0
	}
	m.width = width
}

// View renders the full status bar line.
func (m *Model) View() tea.View {
	status := m.status
	segments := []string{m.renderBrand(), m.renderRuntimeIndicator(status)}
	if organization := m.renderOrganization(status); organization != "" {
		segments = append(segments, organization)
	}

	left := lipgloss.JoinHorizontal(
		lipgloss.Left,
		"  ",
		m.theme.Gradients.Motif.Render(strings.Repeat("╱", 2), false),
		" ",
		strings.Join(segments, " "),
	)
	if m.width <= 0 {
		return tea.NewView(left)
	}

	if lipgloss.Width(left) > m.width {
		available := m.width - lipgloss.Width("  ") - lipgloss.Width(m.theme.Gradients.Motif.Render(strings.Repeat("╱", 2), false)) - lipgloss.Width(m.renderBrand()) - lipgloss.Width(m.renderRuntimeIndicator(status)) - 3
		if available <= 0 {
			return tea.NewView(lipgloss.JoinHorizontal(
				lipgloss.Left,
				"  ",
				m.theme.Gradients.Motif.Render(strings.Repeat("╱", 2), false),
				" ",
				m.renderBrand(),
			))
		}

		organization := truncateLabel(m.renderOrganization(status), available)
		parts := []string{m.renderBrand(), m.renderRuntimeIndicator(status)}
		if organization != "" {
			parts = append(parts, organization)
		}
		left = lipgloss.JoinHorizontal(
			lipgloss.Left,
			"  ",
			m.theme.Gradients.Motif.Render(strings.Repeat("╱", 2), false),
			" ",
			strings.Join(parts, " "),
		)
	}

	fillWidth := m.width - lipgloss.Width(left) - 1
	if fillWidth <= 0 {
		return tea.NewView(left)
	}
	return tea.NewView(lipgloss.JoinHorizontal(
		lipgloss.Left,
		left,
		" ",
		m.theme.Gradients.Motif.Render(strings.Repeat("╱", fillWidth), false),
	))
}

func (m *Model) renderBrand() string {
	return m.theme.Gradients.Brand.Render(strings.ToUpper(theme.AppName), true)
}

func (m *Model) renderOrganization(status accountruntime.Status) string {
	if name := strings.TrimSpace(status.Scope.Organization.Name); name != "" {
		return m.theme.Shell.HeaderLead.Render(name)
	}
	if id := strings.TrimSpace(string(status.Scope.Organization.ID)); id != "" {
		return m.theme.Shell.HeaderLead.Render(id)
	}
	if name := strings.TrimSpace(m.selectedOrganization.Name); name != "" {
		return m.theme.Shell.HeaderLead.Render(name)
	}
	if id := strings.TrimSpace(string(m.selectedOrganization.ID)); id != "" {
		return m.theme.Shell.HeaderLead.Render(id)
	}
	return ""
}

func (m *Model) renderRuntimeIndicator(status accountruntime.Status) string {
	return runtimeIndicatorStyle(m, status).Render("●")
}

func runtimeIndicatorStyle(m *Model, status accountruntime.Status) lipgloss.Style {
	if m.accountSelected && !status.Running {
		return m.theme.Text.Warning
	}
	if !status.Running {
		return m.theme.Text.Muted
	}

	switch status.Sync.(type) {
	case *pssyncer.Ready:
		return m.theme.Text.Success
	case *pssyncer.Connecting, *pssyncer.Reconnecting, *pssyncer.Syncing:
		return m.theme.Text.Warning
	case *pssyncer.Error, *pssyncer.Disconnected:
		return m.theme.Text.Error
	default:
		return m.theme.Text.Muted
	}
}

func truncateLabel(label string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(label)
	if len(runes) <= width {
		return label
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}
