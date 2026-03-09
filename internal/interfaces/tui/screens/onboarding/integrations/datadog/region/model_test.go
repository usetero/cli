package datadogregion

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_SelectEmitsSite(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))

	model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected select command")
	}

	msg := cmd()
	selected, ok := msg.(SelectedMsg)
	if !ok {
		t.Fatalf("expected SelectedMsg, got %T", msg)
	}
	if selected.Site != integrations.DatadogSiteUS3 {
		t.Fatalf("expected site %q, got %q", integrations.DatadogSiteUS3, selected.Site)
	}
}

func TestModel_ViewIncludesRegionDescriptions(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))

	view := model.View().Content
	if !strings.Contains(view, "US1 (datadoghq.com)") {
		t.Fatalf("expected display label in view, got %q", view)
	}
	if strings.Contains(view, "United States") {
		t.Fatalf("did not expect region description rows in compact view, got %q", view)
	}
	if !strings.Contains(view, "Select your Datadog region") {
		t.Fatalf("expected updated title in view, got %q", view)
	}
	if !strings.Contains(view, "Choose the region where your Datadog account is hosted") {
		t.Fatalf("expected explanatory copy in view, got %q", view)
	}
}
