package statusbar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/domains/tenancy"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/events"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
)

func TestModel_RendersBrand(t *testing.T) {
	model := New("dev", theme.New(false))
	view := ansi.Strip(model.View().Content)
	if !strings.Contains(view, "TERO") {
		t.Fatalf("expected brand in status bar, got %q", view)
	}
}

func TestModel_RendersRuntimeIndicator(t *testing.T) {
	tests := []struct {
		name   string
		status accountruntime.Status
	}{
		{
			name:   "offline",
			status: accountruntime.Status{Running: false, Sync: &pssyncer.Disconnected{}},
		},
		{
			name:   "disconnected",
			status: accountruntime.Status{Running: true, Sync: &pssyncer.Disconnected{}},
		},
		{
			name:   "connecting",
			status: accountruntime.Status{Running: true, Sync: &pssyncer.Connecting{}},
		},
		{
			name:   "syncing",
			status: accountruntime.Status{Running: true, Sync: &pssyncer.Syncing{Progress: &pssyncer.Progress{Downloaded: 3, Total: 10}}},
		},
		{
			name:   "ready",
			status: accountruntime.Status{Running: true, Sync: &pssyncer.Ready{}},
		},
		{
			name:   "reconnecting",
			status: accountruntime.Status{Running: true, Sync: &pssyncer.Reconnecting{}},
		},
		{
			name:   "error",
			status: accountruntime.Status{Running: true, Sync: &pssyncer.Error{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := New("dev", theme.New(false))
			model.Update(events.AccountRuntimeUpdatedMsg{Status: tc.status})
			view := ansi.Strip(model.View().Content)
			if !strings.Contains(view, "●") {
				t.Fatalf("expected runtime indicator dot, got %q", view)
			}
		})
	}
}

func TestModel_RendersOrganizationName(t *testing.T) {
	model := New("dev", theme.New(false))
	model.Update(events.OrganizationSelectedMsg{Organization: tenancy.Organization{ID: "org_1", Name: "Acme Corp"}})
	model.SetSize(120, 1)
	view := ansi.Strip(model.View().Content)

	if !strings.Contains(view, "Acme Corp") {
		t.Fatalf("expected organization in status bar, got %q", view)
	}
}

func TestModel_TruncatesOrganizationToFitWidth(t *testing.T) {
	model := New("dev", theme.New(false))
	model.Update(events.OrganizationSelectedMsg{Organization: tenancy.Organization{ID: "org_1", Name: "Acme Corp"}})
	model.SetSize(12, 1)
	view := ansi.Strip(model.View().Content)

	if !strings.Contains(view, "TERO") || !strings.Contains(view, "╱╱") {
		t.Fatalf("expected lead-in and brand to remain visible, got %q", view)
	}
	if strings.Contains(view, "Acme Corp") {
		t.Fatalf("expected truncated organization in narrow mode, got %q", view)
	}
}

func TestModel_UsesWarningDotAfterAccountSelectionBeforeRuntimeStarts(t *testing.T) {
	model := New("dev", theme.New(false))
	model.Update(events.OrganizationSelectedMsg{Organization: tenancy.Organization{ID: "org_1", Name: "Acme Corp"}})
	model.Update(events.AccountSelectedMsg{
		Scope: accountruntime.Scope{
			Organization: tenancy.Organization{ID: "org_1", Name: "Acme Corp"},
			Account:      tenancy.Account{ID: "acc_1", Name: "Main"},
		},
	})

	view := model.View().Content
	if !strings.Contains(view, "Acme Corp") {
		t.Fatalf("expected organization in status bar, got %q", ansi.Strip(view))
	}
}
