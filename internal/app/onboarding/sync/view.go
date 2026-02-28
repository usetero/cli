package sync

import (
	"fmt"

	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/powersync"
)

// View renders the sync UI.
func (m *Model) View() string {
	s := m.theme.Styles
	title := s.Title.Render("Getting ready")

	switch state := m.syncer.State().(type) {
	case *powersync.Ready:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", s.Success.Render("Ready!"))

	case *powersync.Error:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", s.Error.Render(fmt.Sprintf("Error: %v", state.Err)))

	case *powersync.Connecting:
		statusLine := m.spinner.View() + " " + s.Body.Render("Connecting...")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)

	case *powersync.Syncing:
		msg := "Syncing your data..."
		if state.Progress != nil && state.Progress.Total > 0 {
			msg = fmt.Sprintf("Syncing your data... (%s)", state.Progress)
		}
		statusLine := m.spinner.View() + " " + s.Body.Render(msg)
		parts := []string{title, "", statusLine}

		if state.Progress != nil && state.Progress.Total > 0 {
			pct := float64(state.Progress.Downloaded) / float64(state.Progress.Total) * 100
			progressBar := m.progress.ViewAs(pct)
			countText := fmt.Sprintf("%d / %d rows", state.Progress.Downloaded, state.Progress.Total)
			parts = append(parts, "", progressBar, "", s.Help.Render(countText))
		}

		return lipgloss.JoinVertical(lipgloss.Left, parts...)

	case *powersync.Reconnecting:
		statusLine := m.spinner.View() + " " + s.Body.Render("Reconnecting...")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)

	default:
		statusLine := m.spinner.View() + " " + s.Body.Render("Starting...")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)
	}
}
