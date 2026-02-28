package statusbar

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/tea/keymap"
)

// ToggleDrawer toggles the drawer open/closed.
// Opening is suppressed until at least one tab has data.
func (m *Model) ToggleDrawer() {
	if m.drawerOpen {
		m.drawerOpen = false
		return
	}
	if m.anyTabHasData() {
		m.drawerOpen = true
	}
}

// CloseDrawer closes the drawer.
func (m *Model) CloseDrawer() {
	m.drawerOpen = false
}

// NextTab cycles to the next drawer tab.
func (m *Model) NextTab() {
	if len(m.tabs) == 0 {
		return
	}
	m.activeTab = (m.activeTab + 1) % len(m.tabs)
}

// IsDrawerOpen returns whether the drawer is open.
func (m *Model) IsDrawerOpen() bool {
	return m.drawerOpen
}

// HandleKeyPress forwards key events to the active tab's key handler.
func (m *Model) HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	tab := m.activeTabModel()
	if tab == nil || !tab.Interactive() {
		return nil
	}
	return tab.HandleKeyPress(msg)
}

// HandleEsc lets the active tab consume an esc press before the drawer closes.
// Returns true if the tab consumed the event (e.g. back from detail view).
func (m *Model) HandleEsc() bool {
	tab := m.activeTabModel()
	if tab != nil && tab.InDetail() {
		tab.CloseDetail()
		return true
	}
	return false
}

// ShortHelp returns keybindings shown in the keybar when the drawer is open.
func (m *Model) ShortHelp() []key.Binding {
	tab := m.activeTabModel()
	if tab != nil && tab.Interactive() && tab.HasData() {
		if tab.InDetail() {
			return []key.Binding{keymap.DrawerUp, keymap.DrawerDown, keymap.DrawerSelect, keymap.DrawerBack, keymap.NextTab}
		}
		return []key.Binding{keymap.DrawerUp, keymap.DrawerDown, keymap.DrawerSelect, keymap.NextTab, keymap.CloseDrawer}
	}
	return []key.Binding{keymap.NextTab, keymap.CloseDrawer}
}

func (m *Model) activeTabModel() drawerTab {
	if m.activeTab < 0 || m.activeTab >= len(m.tabs) {
		return nil
	}
	return m.tabs[m.activeTab]
}

func (m *Model) anyTabHasData() bool {
	for _, tab := range m.tabs {
		if tab.HasData() {
			return true
		}
	}
	return false
}
