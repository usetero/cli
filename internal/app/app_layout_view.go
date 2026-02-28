package app

import (
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/tea/cursor"
	"github.com/usetero/cli/internal/tea/keymap"
)

type renderFrame struct {
	paddedView      string
	contentWidth    int
	pageHeight      int
	toastHeight     int
	statusBarHeight int
}

// updateLayout propagates sizes to children based on current dimensions.
func (m *Model) updateLayout() {
	contentWidth, contentHeight := m.contentSize()

	// Fixed components get width, report their height
	m.statusBar.SetWidth(contentWidth)
	m.toast.SetWidth(contentWidth)
	m.keyBar.SetWidth(contentWidth)

	statusBarHeight := m.statusBar.Height()
	toastHeight := m.toast.Height()
	keyBarHeight := m.keyBar.Height()

	// Page is flexible - gets remaining height
	// Toast always reserves its line to prevent layout shifts
	pageHeight := contentHeight - statusBarHeight - gapAfterStatusBar - toastHeight - gapBeforeKeyBar - keyBarHeight

	switch m.state {
	case stateOnboarding:
		if m.onboarding != nil {
			m.onboarding.SetSize(contentWidth, pageHeight)
		}
	case stateChat:
		if m.chat != nil {
			m.chat.SetSize(contentWidth, pageHeight)
			// Chat page origin: toast + statusbar + gap (no top padding)
			m.chat.SetOrigin(horizontalPadding, toastHeight+statusBarHeight+gapAfterStatusBar)
		}
	}
}

// contentSize returns the available space for content (after app padding).
func (m *Model) contentSize() (int, int) {
	if m.width == 0 || m.height == 0 {
		return 0, 0
	}
	contentWidth := m.width - (horizontalPadding * 2)
	contentHeight := m.height - verticalPadding // bottom padding only, no top padding
	return contentWidth, contentHeight
}

// updateKeyBar updates the keybar with current page bindings plus global bindings.
func (m *Model) updateKeyBar() {
	var bindings []key.Binding

	if m.statusBar.IsDrawerOpen() {
		bindings = m.statusBar.ShortHelp()
	} else {
		switch m.state {
		case stateOnboarding:
			if m.onboarding != nil {
				bindings = m.onboarding.ShortHelp()
			}
		case stateChat:
			if m.chat != nil {
				bindings = m.chat.ShortHelp()
			}
		}
	}

	// Always append global bindings
	bindings = append(bindings, keymap.Global...)
	m.keyBar.SetKeyBindings(bindings)
}

// View renders the app.
func (m *Model) View() tea.View {
	start := time.Now()
	defer m.logSlowRender(start)

	colors := m.theme

	// Show message if window is too small
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

	// Render content
	rendered := m.renderContent()

	// Extract cursor marker and set cursor position
	cleanView, cur := cursor.Extract(rendered)
	if cur != nil {
		cur.Color = colors.Accent
	}

	// Suppress cursor when drawer or quit dialog is open (palette keeps its cursor)
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
// Layout: statusbar | page | toast (optional) | keybar
func (m *Model) renderContent() string {
	frame, ok := m.buildRenderFrame()
	if !ok {
		return ""
	}

	// Overlay drawer if open
	if m.statusBar.IsDrawerOpen() {
		return m.overlayDrawer(frame)
	}

	// Overlay palette if open
	if m.palette != nil {
		return m.renderPaletteOverlay(frame.paddedView)
	}

	// Overlay quit dialog if open
	if m.quitDlg != nil {
		return m.overlayQuitDialog(frame.paddedView)
	}

	return frame.paddedView
}

// renderPaletteOverlay renders the palette centered on screen.
func (m *Model) renderPaletteOverlay(base string) string {
	paletteView := m.palette.View()
	paletteW := lipgloss.Width(paletteView)
	paletteH := lipgloss.Height(paletteView)
	centerX := (m.width - paletteW) / 2
	centerY := (m.height - paletteH) / 2

	// Extract cursor marker before compositing (compositor strips OSC sequences)
	cleanPalette, paletteCur := cursor.Extract(paletteView)

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(base).X(0).Y(0),
		lipgloss.NewLayer(cleanPalette).X(centerX).Y(centerY),
	}
	result := lipgloss.NewCompositor(layers...).Render()

	// Re-insert cursor marker at the composited position
	if paletteCur != nil {
		result = cursor.Insert(result, paletteCur.X+centerX, paletteCur.Y+centerY)
	}

	return result
}

func (m *Model) buildRenderFrame() (renderFrame, bool) {
	if m.width == 0 || m.height == 0 {
		return renderFrame{}, false
	}

	contentWidth, contentHeight := m.contentSize()

	statusBarView := m.statusBar.View()
	statusBarHeight := m.statusBar.Height()
	toastView := m.toast.View()
	toastHeight := m.toast.Height()
	keyBarView := m.keyBar.View()
	keyBarHeight := m.keyBar.Height()

	pageHeight := contentHeight - statusBarHeight - gapAfterStatusBar - toastHeight - gapBeforeKeyBar - keyBarHeight

	styledPage := lipgloss.NewStyle().
		Width(contentWidth).
		Height(pageHeight).
		MaxHeight(pageHeight).
		Render(m.currentPageView())

	var sections []string
	sections = append(sections, toastView)
	sections = append(sections, statusBarView)
	for range gapAfterStatusBar {
		sections = append(sections, "")
	}
	sections = append(sections, styledPage)
	for range gapBeforeKeyBar {
		sections = append(sections, "")
	}
	if keyBarHeight > 0 {
		sections = append(sections, keyBarView)
	}

	innerView := lipgloss.JoinVertical(lipgloss.Left, sections...)
	paddedView := lipgloss.NewStyle().
		PaddingTop(0).
		PaddingRight(horizontalPadding).
		PaddingBottom(verticalPadding).
		PaddingLeft(horizontalPadding).
		Render(innerView)

	return renderFrame{
		paddedView:      paddedView,
		contentWidth:    contentWidth,
		pageHeight:      pageHeight,
		toastHeight:     toastHeight,
		statusBarHeight: statusBarHeight,
	}, true
}

func (m *Model) currentPageView() string {
	switch m.state {
	case stateOnboarding:
		if m.onboarding != nil {
			return m.onboarding.View()
		}
	case stateChat:
		if m.chat != nil {
			return m.chat.View()
		}
	}
	return ""
}

func (m *Model) overlayDrawer(frame renderFrame) string {
	drawerWidth := frame.contentWidth - 2
	drawerHeight := frame.pageHeight - 2 // fill page area, leave gap at bottom
	if drawerHeight < 6 {
		drawerHeight = 6
	}
	drawer := m.statusBar.DrawerView(drawerWidth, drawerHeight)

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(frame.paddedView).X(0).Y(0),
		lipgloss.NewLayer(drawer).X(horizontalPadding + 1).Y(frame.toastHeight + frame.statusBarHeight),
	}
	return lipgloss.NewCompositor(layers...).Render()
}

func (m *Model) overlayQuitDialog(base string) string {
	dialog := m.quitDlg.View()
	dialogW := lipgloss.Width(dialog)
	dialogH := lipgloss.Height(dialog)
	centerX := (m.width - dialogW) / 2
	centerY := (m.height - dialogH) / 2

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(base).X(0).Y(0),
		lipgloss.NewLayer(dialog).X(centerX).Y(centerY),
	}
	return lipgloss.NewCompositor(layers...).Render()
}
