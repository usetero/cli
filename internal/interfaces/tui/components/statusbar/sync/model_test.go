package sync

import (
	"testing"

	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/events"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
)

func TestModel_MapsRuntimeStatusToDisplayState(t *testing.T) {
	tests := []struct {
		name   string
		status accountruntime.Status
		want   state
	}{
		{
			name: "default status stays muted",
			want: stateMuted,
		},
		{
			name:   "connecting",
			status: accountruntime.Status{Sync: &pssyncer.Connecting{}},
			want:   stateConnecting,
		},
		{
			name:   "ready",
			status: accountruntime.Status{Sync: &pssyncer.Ready{}},
			want:   stateReady,
		},
		{
			name:   "reconnecting",
			status: accountruntime.Status{Sync: &pssyncer.Reconnecting{}},
			want:   stateReconnecting,
		},
		{
			name:   "syncing",
			status: accountruntime.Status{Sync: &pssyncer.Syncing{}},
			want:   stateSyncing,
		},
		{
			name:   "disconnected",
			status: accountruntime.Status{Sync: &pssyncer.Disconnected{}},
			want:   stateError,
		},
		{
			name:   "error",
			status: accountruntime.Status{Sync: &pssyncer.Error{}},
			want:   stateError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := New(theme.New(false))
			model.Update(events.AccountRuntimeUpdatedMsg{Status: tc.status})
			if model.state != tc.want {
				t.Fatalf("expected state %v, got %v", tc.want, model.state)
			}
		})
	}
}
