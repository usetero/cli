package statusbar

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/styles"
)

func TestToggleDrawerRequiresData(t *testing.T) {
	t.Parallel()

	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com", "dev")
	m.tabs = []drawerTab{
		drawerTabAdapter{label: "A", hasData: func() bool { return false }},
		drawerTabAdapter{label: "B", hasData: func() bool { return true }},
	}

	m.ToggleDrawer()
	if !m.drawerOpen {
		t.Fatalf("drawer should open when at least one tab has data")
	}

	m.ToggleDrawer()
	if m.drawerOpen {
		t.Fatalf("drawer should close when toggled while open")
	}
}

func TestHandleEscDelegatesToActiveTab(t *testing.T) {
	t.Parallel()

	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com", "dev")
	closed := false
	m.tabs = []drawerTab{
		drawerTabAdapter{
			label:       "A",
			inDetail:    func() bool { return true },
			closeDetail: func() { closed = true },
		},
	}
	m.activeTab = 0

	if !m.HandleEsc() {
		t.Fatalf("expected esc to be consumed by active detail tab")
	}
	if !closed {
		t.Fatalf("expected esc to close active tab detail")
	}
}

func TestHandleKeyPressUsesInteractiveTabsOnly(t *testing.T) {
	t.Parallel()

	m := New(styles.NewTheme(true), logtest.NewScope(t), powersynctest.NewMockSyncer(), "https://api.example.com", "dev")
	called := false
	m.tabs = []drawerTab{
		drawerTabAdapter{
			label:       "A",
			interactive: false,
			handleKey: func(msg tea.KeyPressMsg) tea.Cmd {
				called = true
				return nil
			},
		},
		drawerTabAdapter{
			label:       "B",
			interactive: true,
			handleKey: func(msg tea.KeyPressMsg) tea.Cmd {
				called = true
				return nil
			},
		},
	}

	m.activeTab = 0
	_ = m.HandleKeyPress(tea.KeyPressMsg{})
	if called {
		t.Fatalf("non-interactive tab should not receive key presses")
	}

	m.activeTab = 1
	_ = m.HandleKeyPress(tea.KeyPressMsg{})
	if !called {
		t.Fatalf("interactive tab should receive key presses")
	}
}
