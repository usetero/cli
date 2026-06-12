package statusbar

import (
	"strings"
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func TestRenderOrgContext(t *testing.T) {
	m := New(styles.NewTheme(true), logtest.NewScope(t), "https://api.example.com", "dev")

	m.org = "Acme"
	view := m.renderOrgWorkspace()
	if !strings.Contains(view, "Acme") {
		t.Fatalf("expected org in view, got %q", view)
	}
}

func TestBuildTabsMirrorsProductSurfaces(t *testing.T) {
	m := New(styles.NewTheme(true), logtest.NewScope(t), "https://api.example.com", "dev")

	want := []struct {
		group string
		label string
	}{
		{group: "Control Plane", label: "Issues"},
		{group: "Control Plane", label: "Checks"},
		{group: "Data Plane", label: "Services"},
		{group: "Data Plane", label: "Log events"},
		{group: "Data Plane", label: "Edge instances"},
	}

	if len(m.tabs) != len(want) {
		t.Fatalf("tab count = %d, want %d", len(m.tabs), len(want))
	}
	for i, tab := range m.tabs {
		if tab.Label() != want[i].label {
			t.Fatalf("tab %d label = %q, want %q", i, tab.Label(), want[i].label)
		}
		grouped, ok := tab.(groupedDrawerTab)
		if !ok {
			t.Fatalf("tab %d does not expose a group", i)
		}
		if grouped.GroupLabel() != want[i].group {
			t.Fatalf("tab %d group = %q, want %q", i, grouped.GroupLabel(), want[i].group)
		}
	}
}
