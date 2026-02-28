package datadog

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func TestRegionEnterEmitsSelectedSite(t *testing.T) {
	t.Parallel()

	m := NewRegion(styles.NewTheme(true), logtest.NewScope(t))
	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected selection command")
	}

	msg := cmd()
	selected, ok := msg.(msgs.DatadogRegionSelected)
	if !ok {
		t.Fatalf("expected DatadogRegionSelected, got %T", msg)
	}
	if selected.Site != domain.DatadogRegions[0].Site {
		t.Fatalf("expected default selected site %q, got %q", domain.DatadogRegions[0].Site, selected.Site)
	}
}

func TestRegionNavigationChangesSelection(t *testing.T) {
	t.Parallel()

	m := NewRegion(styles.NewTheme(true), logtest.NewScope(t))
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected selection command")
	}
	msg := cmd()
	selected, ok := msg.(msgs.DatadogRegionSelected)
	if !ok {
		t.Fatalf("expected DatadogRegionSelected, got %T", msg)
	}
	if len(domain.DatadogRegions) < 2 {
		t.Fatal("test requires at least two datadog regions")
	}
	if selected.Site != domain.DatadogRegions[1].Site {
		t.Fatalf("expected second site %q, got %q", domain.DatadogRegions[1].Site, selected.Site)
	}
}
