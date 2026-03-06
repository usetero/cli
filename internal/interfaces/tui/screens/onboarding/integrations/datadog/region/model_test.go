package datadogregion

import (
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
