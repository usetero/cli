package chat

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
)

// View renders the chat. This is a flexible component - renders exactly to SetSize dimensions.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Empty state: context-aware greeting + suggestions + input bar.
	if !m.hasMessages() {
		emptyHeight := m.height - m.inputBar.Height()
		emptyView := lipgloss.NewStyle().
			Width(m.width).
			Height(emptyHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Render(m.emptyStateContent())

		return lipgloss.JoinVertical(lipgloss.Left, emptyView, m.inputBar.View())
	}

	// Normal state: message list + spacer + input bar.
	spacer := lipgloss.NewStyle().Width(m.width).Background(m.theme.Bg).Render("")
	return lipgloss.JoinVertical(lipgloss.Left, m.messageList.View(), spacer, m.inputBar.View())
}

// emptyStateContent renders the context-aware empty state.
func (m *Model) emptyStateContent() string {
	colors := m.theme
	text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)

	name := ""
	if m.user != nil && m.user.FirstName != "" {
		name = m.user.FirstName
	}

	var headline string
	var suggestions []string

	s := m.policySummary
	if s == nil {
		// Still loading.
		if name != "" {
			headline = fmt.Sprintf("Hey %s — loading your environment...", name)
		} else {
			headline = "Loading your environment..."
		}
	} else if s.ActiveServices == 0 {
		// No services enabled — guide them to get started.
		if name != "" {
			headline = fmt.Sprintf("Welcome, %s — let's get your environment set up.", name)
		} else {
			headline = "Let's get your environment set up."
		}
		suggestions = []string{
			"Help me get started",
			"What services do I have?",
			"What does Tero do?",
		}
	} else if s.PendingPolicyCount > 0 {
		// Pending work.
		wp := wastePercent(*s)
		if name != "" {
			headline = fmt.Sprintf("Hey %s — %d%% waste across your services. Let's dig in:", name, wp)
		} else {
			headline = fmt.Sprintf("%d%% waste across your services. Let's dig in:", wp)
		}
		suggestions = []string{
			"Walk me through the pending policies",
			"Which policies should I approve first?",
			"Show me what's driving the most waste",
		}
	} else if s.ApprovedPolicyCount > 0 {
		// Mid-journey — approved some, none pending.
		if name != "" {
			headline = fmt.Sprintf("Nice work, %s — %d policies approved. Let's keep going:", name, s.ApprovedPolicyCount)
		} else {
			headline = fmt.Sprintf("%d policies approved. Let's keep going:", s.ApprovedPolicyCount)
		}
		suggestions = []string{
			"How are the approved policies performing?",
			"Are there any new recommendations?",
			"Show me observed savings so far",
		}
	} else {
		// All clean — no pending, no approved.
		if name != "" {
			headline = fmt.Sprintf("Looking good, %s — I'm watching for changes.", name)
		} else {
			headline = "Looking good — I'm watching for changes."
		}
		suggestions = []string{
			"What services are generating the most logs?",
			"Show me a summary of my environment",
			"Any optimization opportunities?",
		}
	}

	var lines []string
	lines = append(lines, text.Render(headline))
	lines = append(lines, "")
	for _, s := range suggestions {
		lines = append(lines, muted.Render("  → "+s))
	}

	return strings.Join(lines, "\n")
}

// wastePercent computes waste % preferring bytes.
func wastePercent(s domain.AccountSummary) int {
	if s.TotalBytesPerHour != nil && *s.TotalBytesPerHour > 0 &&
		s.EstimatedBytesPerHour != nil && *s.EstimatedBytesPerHour > 0 {
		return int(math.Round(*s.EstimatedBytesPerHour / *s.TotalBytesPerHour * 100))
	}
	if s.EstimatedCostPerHour != nil && s.TotalCostPerHour != nil && *s.TotalCostPerHour > 0 {
		return int(math.Round(*s.EstimatedCostPerHour / *s.TotalCostPerHour * 100))
	}
	if s.TotalVolumePerHour != nil && *s.TotalVolumePerHour > 0 &&
		s.EstimatedVolumePerHour != nil && *s.EstimatedVolumePerHour > 0 {
		return int(math.Round(*s.EstimatedVolumePerHour / *s.TotalVolumePerHour * 100))
	}
	return 0
}
