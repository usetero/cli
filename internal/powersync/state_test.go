package powersync_test

import (
	"testing"

	"github.com/usetero/cli/internal/powersync"
)

func TestSyncing_UpdateProgress(t *testing.T) {
	t.Parallel()

	t.Run("preserves warning", func(t *testing.T) {
		t.Parallel()

		state := powersync.NewSyncing("Syncing...").WithWarning("Could not apply checkpoint")

		updated := state.UpdateProgress(50, 100)

		if updated.Warning != "Could not apply checkpoint" {
			t.Errorf("Warning = %q, want %q", updated.Warning, "Could not apply checkpoint")
		}
	})

	t.Run("updates message with progress", func(t *testing.T) {
		t.Parallel()

		state := powersync.NewSyncing("Syncing...")

		updated := state.UpdateProgress(50, 100)

		want := "Syncing your data... (50/100)"
		if updated.Message != want {
			t.Errorf("Message = %q, want %q", updated.Message, want)
		}
	})

	t.Run("sets progress", func(t *testing.T) {
		t.Parallel()

		state := powersync.NewSyncing("Syncing...")

		updated := state.UpdateProgress(50, 100)

		if updated.Progress == nil {
			t.Fatal("Progress should not be nil")
		}
		if updated.Progress.Downloaded != 50 {
			t.Errorf("Downloaded = %d, want 50", updated.Progress.Downloaded)
		}
		if updated.Progress.Total != 100 {
			t.Errorf("Total = %d, want 100", updated.Progress.Total)
		}
	})
}

func TestSyncing_WithWarning(t *testing.T) {
	t.Parallel()

	t.Run("preserves message and progress", func(t *testing.T) {
		t.Parallel()

		state := powersync.NewSyncing("Syncing...").WithProgress(50, 100)

		updated := state.WithWarning("Something happened")

		if updated.Message != "Syncing..." {
			t.Errorf("Message = %q, want %q", updated.Message, "Syncing...")
		}
		if updated.Progress == nil || updated.Progress.Downloaded != 50 {
			t.Error("Progress should be preserved")
		}
		if updated.Warning != "Something happened" {
			t.Errorf("Warning = %q, want %q", updated.Warning, "Something happened")
		}
	})
}
