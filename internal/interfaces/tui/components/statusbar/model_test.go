package statusbar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/domains/tenancy"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	sessionruntime "github.com/usetero/cli/internal/runtime/session"
)

type sessionStub struct {
	status sessionruntime.Status
}

func (s sessionStub) Status() sessionruntime.Status { return s.status }

func TestModel_BrandAndEnv(t *testing.T) {
	model := New(sessionStub{}, "dev", theme.New(false))
	view := ansi.Strip(model.View())
	if !strings.Contains(view, "TERO") {
		t.Fatalf("expected brand in status bar, got %q", view)
	}
	if !strings.Contains(view, "DEV") {
		t.Fatalf("expected non-prod env badge in status bar, got %q", view)
	}
}

func TestModel_HidesProdEnv(t *testing.T) {
	model := New(sessionStub{}, "prd", theme.New(false))
	view := ansi.Strip(model.View())
	if strings.Contains(view, "PRD") {
		t.Fatalf("did not expect prod env badge, got %q", view)
	}
}

func TestModel_SyncStates(t *testing.T) {
	tests := []struct {
		name   string
		status sessionruntime.Status
		want   string
	}{
		{
			name:   "offline",
			status: sessionruntime.Status{Running: false, Sync: &pssyncer.Disconnected{}},
			want:   "●",
		},
		{
			name:   "disconnected",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Disconnected{}},
			want:   "●",
		},
		{
			name:   "connecting",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Connecting{}},
			want:   "●",
		},
		{
			name:   "syncing",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Syncing{Progress: &pssyncer.Progress{Downloaded: 3, Total: 10}}},
			want:   "●",
		},
		{
			name:   "ready",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Ready{}},
			want:   "●",
		},
		{
			name:   "reconnecting",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Reconnecting{}},
			want:   "●",
		},
		{
			name:   "error",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Error{}},
			want:   "●",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := New(sessionStub{status: tc.status}, "dev", theme.New(false))
			view := ansi.Strip(model.View())
			if !strings.Contains(view, tc.want) {
				t.Fatalf("expected %q in status bar, got %q", tc.want, view)
			}
			if strings.Contains(view, "offline") || strings.Contains(view, "ready") || strings.Contains(view, "connecting") {
				t.Fatalf("did not expect textual sync state in visual header, got %q", view)
			}
		})
	}
}

func TestModel_RendersSessionContextAndHint(t *testing.T) {
	model := New(
		sessionStub{
			status: sessionruntime.Status{
				Running: true,
				Sync:    &pssyncer.Ready{},
				Scope: sessionruntime.Scope{
					Organization: tenancy.Organization{ID: "org_1", Name: "Acme Corp"},
					Account:      tenancy.Account{ID: "acc_1", Name: "Main"},
					Workspace:    tenancy.Workspace{ID: "ws_1", Name: "Production", AccountID: "acc_1"},
				},
			},
		},
		"dev",
		theme.New(false),
	)
	model.SetWidth(120)
	view := ansi.Strip(model.View())

	if !strings.Contains(view, "Acme Corp") {
		t.Fatalf("expected session context in status bar, got %q", view)
	}
	if strings.Contains(view, "Production") {
		t.Fatalf("did not expect workspace label without disambiguation rule, got %q", view)
	}
	if !strings.Contains(view, "ctrl+d open") {
		t.Fatalf("expected drawer hint in status bar, got %q", view)
	}
	if strings.Contains(view, "ready") {
		t.Fatalf("did not expect textual ready label when context is available, got %q", view)
	}
	if strings.Contains(view, "DEV") {
		t.Fatalf("did not expect env label when organization context is available, got %q", view)
	}
}

func TestModel_WidthAwareDegradesBeforeDroppingContext(t *testing.T) {
	model := New(
		sessionStub{
			status: sessionruntime.Status{
				Running: true,
				Sync:    &pssyncer.Ready{},
				Scope: sessionruntime.Scope{
					Organization: tenancy.Organization{ID: "org_1", Name: "Acme Corp"},
					Account:      tenancy.Account{ID: "acc_1", Name: "Main"},
					Workspace:    tenancy.Workspace{ID: "ws_1", Name: "Production", AccountID: "acc_1"},
				},
			},
		},
		"development",
		theme.New(false),
	)
	model.SetWidth(36)
	view := ansi.Strip(model.View())

	if strings.Contains(view, "DEVELOPMENT") {
		t.Fatalf("expected env badge to drop in narrow mode, got %q", view)
	}
	if strings.Contains(view, "ctrl+d open") {
		t.Fatalf("expected drawer hint to drop before core context, got %q", view)
	}
	if !strings.Contains(view, "Acme Corp") {
		t.Fatalf("expected organization context to remain visible, got %q", view)
	}
	if strings.Contains(view, "ready") {
		t.Fatalf("did not expect sync text fallback when context exists, got %q", view)
	}
}
