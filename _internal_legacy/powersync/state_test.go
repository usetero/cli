package powersync_test

import (
	"testing"

	"github.com/usetero/cli/internal/powersync"
)

func TestSyncing_WithProgress(t *testing.T) {
	t.Parallel()

	t.Run("sets progress", func(t *testing.T) {
		t.Parallel()

		state := powersync.NewSyncing().WithProgress(50, 100)

		if state.Progress == nil {
			t.Fatal("Progress should not be nil")
		}
		if state.Progress.Downloaded != 50 {
			t.Errorf("Downloaded = %d, want 50", state.Progress.Downloaded)
		}
		if state.Progress.Total != 100 {
			t.Errorf("Total = %d, want 100", state.Progress.Total)
		}
	})
}

func TestProgress_String(t *testing.T) {
	t.Parallel()

	p := &powersync.Progress{Downloaded: 50, Total: 100}
	want := "50/100"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestReconnecting(t *testing.T) {
	t.Parallel()

	t.Run("not degraded", func(t *testing.T) {
		t.Parallel()

		state := powersync.NewReconnecting(false)
		if state.Degraded {
			t.Error("expected Degraded to be false")
		}
	})

	t.Run("degraded", func(t *testing.T) {
		t.Parallel()

		state := powersync.NewReconnecting(true)
		if !state.Degraded {
			t.Error("expected Degraded to be true")
		}
	})
}
