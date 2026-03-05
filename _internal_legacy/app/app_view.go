package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/tea/cursor"
)

// View renders the app.
func (m *Model) View() tea.View {
	start := time.Now()
	defer m.logSlowRender(start)

	colors := m.theme

	// Show message if window is too small.
	if m.width < minWidth || m.height < minHeight {
		content := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(
				lipgloss.NewStyle().
					Padding(0, 2).
					Foreground(colors.Text).
					BorderStyle(lipgloss.RoundedBorder()).
					BorderForeground(colors.Accent).
					Render("Window too small"),
			)
		return tea.View{
			Content:         content,
			BackgroundColor: colors.Bg,
			AltScreen:       true,
			WindowTitle:     m.windowTitle,
		}
	}

	rendered := m.renderContent()

	// Extract cursor marker and set cursor position.
	cleanView, cur := cursor.Extract(rendered)
	if cur != nil {
		cur.Color = colors.Accent
	}

	// Suppress cursor when drawer or quit dialog is open (palette keeps its cursor).
	if m.statusBar.IsDrawerOpen() || m.quitDlg != nil {
		cur = nil
	}

	return tea.View{
		Content:         cleanView,
		BackgroundColor: colors.Bg,
		AltScreen:       true,
		Cursor:          cur,
		MouseMode:       tea.MouseModeCellMotion,
		WindowTitle:     m.windowTitle,
	}
}

// renderContent renders the main content with padding.
// Layout: statusbar | page | toast (optional) | keybar.
func (m *Model) renderContent() string {
	frame, ok := m.buildRenderFrame()
	if !ok {
		return ""
	}

	if m.statusBar.IsDrawerOpen() {
		return m.overlayDrawer(frame)
	}
	if m.palette != nil {
		return m.renderPaletteOverlay(frame.paddedView)
	}
	if m.quitDlg != nil {
		return m.overlayQuitDialog(frame.paddedView)
	}
	return frame.paddedView
}
