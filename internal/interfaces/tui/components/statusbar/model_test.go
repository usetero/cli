package statusbar

import (
	"strings"
	"testing"

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
	view := model.View()
	if !strings.Contains(view, "TERO") {
		t.Fatalf("expected brand in status bar, got %q", view)
	}
	if !strings.Contains(view, "DEV") {
		t.Fatalf("expected non-prod env badge in status bar, got %q", view)
	}
}

func TestModel_HidesProdEnv(t *testing.T) {
	model := New(sessionStub{}, "prd", theme.New(false))
	view := model.View()
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
			want:   "offline",
		},
		{
			name:   "disconnected",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Disconnected{}},
			want:   "disconnected",
		},
		{
			name:   "connecting",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Connecting{}},
			want:   "connecting",
		},
		{
			name:   "syncing",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Syncing{Progress: &pssyncer.Progress{Downloaded: 3, Total: 10}}},
			want:   "sync 3/10",
		},
		{
			name:   "ready",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Ready{}},
			want:   "ready",
		},
		{
			name:   "reconnecting",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Reconnecting{}},
			want:   "reconnecting",
		},
		{
			name:   "error",
			status: sessionruntime.Status{Running: true, Sync: &pssyncer.Error{}},
			want:   "error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := New(sessionStub{status: tc.status}, "dev", theme.New(false))
			view := model.View()
			if !strings.Contains(view, tc.want) {
				t.Fatalf("expected %q in status bar, got %q", tc.want, view)
			}
		})
	}
}

func TestModel_WidthAwareDegradesToCompact(t *testing.T) {
	model := New(
		sessionStub{
			status: sessionruntime.Status{
				Running: true,
				Sync:    &pssyncer.Connecting{},
			},
		},
		"development",
		theme.New(false),
	)
	model.SetWidth(18)
	view := model.View()

	if strings.Contains(view, "DEVELOPMENT") {
		t.Fatalf("expected env badge to drop in narrow mode, got %q", view)
	}
	if !strings.Contains(view, "conn") {
		t.Fatalf("expected compact sync label in narrow mode, got %q", view)
	}
}
