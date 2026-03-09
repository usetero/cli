package statusbar

import (
	"testing"

	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	sessionruntime "github.com/usetero/cli/internal/runtime/session"
)

func TestPresentSync(t *testing.T) {
	tests := []struct {
		name      string
		status    sessionruntime.Status
		compact   bool
		wantIcon  string
		wantLabel string
		wantTone  syncTone
	}{
		{
			name:      "offline full",
			status:    sessionruntime.Status{Running: false, Sync: &pssyncer.Disconnected{}},
			compact:   false,
			wantIcon:  "●",
			wantLabel: "offline",
			wantTone:  syncToneError,
		},
		{
			name:      "offline compact",
			status:    sessionruntime.Status{Running: false, Sync: &pssyncer.Disconnected{}},
			compact:   true,
			wantIcon:  "●",
			wantLabel: "off",
			wantTone:  syncToneError,
		},
		{
			name:      "syncing progress full",
			status:    sessionruntime.Status{Running: true, Sync: &pssyncer.Syncing{Progress: &pssyncer.Progress{Downloaded: 2, Total: 7}}},
			compact:   false,
			wantIcon:  "●",
			wantLabel: "sync 2/7",
			wantTone:  syncToneWarning,
		},
		{
			name:      "connecting compact",
			status:    sessionruntime.Status{Running: true, Sync: &pssyncer.Connecting{}},
			compact:   true,
			wantIcon:  "●",
			wantLabel: "conn",
			wantTone:  syncToneWarning,
		},
		{
			name:      "ready",
			status:    sessionruntime.Status{Running: true, Sync: &pssyncer.Ready{}},
			compact:   false,
			wantIcon:  "●",
			wantLabel: "ready",
			wantTone:  syncToneSuccess,
		},
		{
			name:      "error compact",
			status:    sessionruntime.Status{Running: true, Sync: &pssyncer.Error{}},
			compact:   true,
			wantIcon:  "●",
			wantLabel: "err",
			wantTone:  syncToneError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := presentSync(tc.status, tc.compact)
			if got.icon != tc.wantIcon {
				t.Fatalf("expected icon %q, got %q", tc.wantIcon, got.icon)
			}
			if got.label != tc.wantLabel {
				t.Fatalf("expected label %q, got %q", tc.wantLabel, got.label)
			}
			if got.tone != tc.wantTone {
				t.Fatalf("expected tone %v, got %v", tc.wantTone, got.tone)
			}
		})
	}
}
