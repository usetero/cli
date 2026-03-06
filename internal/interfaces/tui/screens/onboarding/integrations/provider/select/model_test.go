package providerselect

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_SelectProvider(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	model.SetProviders([]integrations.Provider{
		integrations.ProviderDatadog,
	})

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected select command")
	}
	msg := cmd()
	selected, ok := msg.(SelectedMsg)
	if !ok {
		t.Fatalf("expected SelectedMsg, got %T", msg)
	}
	if selected.Provider != integrations.ProviderDatadog {
		t.Fatalf("expected provider %q, got %q", integrations.ProviderDatadog, selected.Provider)
	}
}
