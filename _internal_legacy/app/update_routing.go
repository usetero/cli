package app

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	msgs "github.com/usetero/cli/internal/app/chat/events"
	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/tea/keymap"
)

func (m *Model) handleGlobalMessage(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case quitConfirmed:
		m.shutdown()
		return tea.Quit, true
	case quitCanceled:
		m.quitDlg = nil
		return nil, true
	case appevents.PaletteOpenRequested:
		return m.openPalette(), true
	case appevents.PaletteCloseRequested:
		m.palette = nil
		return nil, true
	case appevents.ThemeChangeRequested:
		return m.setTheme(msg.Theme), true
	default:
		return nil, false
	}
}

func (m *Model) handleInteractionMessage(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return nil, true
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	case tea.MouseClickMsg:
		if m.statusBar.IsDrawerOpen() {
			m.statusBar.CloseDrawer()
		}
		// Let downstream components observe mouse messages.
		return nil, false
	case appevents.DrawerPromptRequested:
		m.statusBar.CloseDrawer()
		return func() tea.Msg { return msgs.UserSubmittedInput{Text: msg.Text} }, true
	case msgs.UserSubmittedInput:
		text := strings.TrimSpace(msg.Text)
		if strings.EqualFold(text, "exit") || strings.EqualFold(text, "quit") {
			m.quitDlg = newQuitDialog(m.theme)
			return nil, true
		}
		return nil, false
	default:
		return nil, false
	}
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	// ctrl+c always quits immediately, regardless of overlays.
	if key.Matches(msg, keymap.Quit) {
		m.shutdown()
		return tea.Quit, true
	}

	// When quit dialog is open, forward keys to it and consume.
	if m.quitDlg != nil {
		return m.quitDlg.Update(msg), true
	}

	// When palette is open, forward keys to it and consume.
	if m.palette != nil {
		return m.palette.Update(msg), true
	}

	if key.Matches(msg, keymap.Exit) {
		if m.statusBar.IsDrawerOpen() {
			// Let active tab handle esc first (e.g. back from detail view).
			if m.statusBar.HandleEsc() {
				return nil, true
			}
			m.statusBar.CloseDrawer()
			return nil, true
		}
		// Esc cancels active round first; only show dialog if nothing to cancel.
		if m.chat != nil {
			if cancelled, cmd := m.chat.CancelActiveRound(); cancelled {
				return cmd, true
			}
		}
		m.quitDlg = newQuitDialog(m.theme)
		return nil, true
	}

	if key.Matches(msg, keymap.Details) {
		m.statusBar.ToggleDrawer()
		return nil, true
	}

	if m.statusBar.IsDrawerOpen() {
		if key.Matches(msg, keymap.Tab) {
			m.statusBar.NextTab()
			return nil, true
		}
		// Forward to active tab for navigation (up/down/enter).
		return m.statusBar.HandleKeyPress(msg), true
	}

	return nil, false
}

func (m *Model) updateChildren(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Palette needs non-key messages (e.g. blink ticks) when open.
	if m.palette != nil {
		cmds = append(cmds, m.palette.Update(msg))
	}

	if cmd := m.statusBar.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if cmd := m.toast.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	switch m.state {
	case stateOnboarding:
		if m.onboarding != nil {
			cmds = append(cmds, m.onboarding.Update(msg))
		}
	case stateChat:
		if m.chat != nil {
			cmds = append(cmds, m.chat.Update(msg))
		}
	}

	// Keep keybar synchronized after page updates.
	m.updateKeyBar()
	return tea.Batch(cmds...)
}
