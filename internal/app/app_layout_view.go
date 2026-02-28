package app

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
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
