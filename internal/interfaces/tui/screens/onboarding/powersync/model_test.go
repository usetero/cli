package powersyncready

import (
	"errors"
	"testing"

	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
)

func TestModelStatusAndProgress(t *testing.T) {
	t.Parallel()

	t.Run("default status", func(t *testing.T) {
		t.Parallel()

		m := New(theme.New(false))
		if got := m.statusLine(); got != "Starting the local sync engine." {
			t.Fatalf("statusLine() = %q", got)
		}
		if detail := m.detailLine(); detail != "" {
			t.Fatalf("detailLine() = %q, want empty", detail)
		}
		if _, ok := m.percent(); ok {
			t.Fatalf("percent() should be unavailable by default")
		}
	})

	t.Run("syncing exposes progress", func(t *testing.T) {
		t.Parallel()

		m := New(theme.New(false))
		m.SetStatus(accountruntime.Status{
			Sync: &pssyncer.Syncing{
				Progress: &pssyncer.Progress{
					Downloaded: 25,
					Total:      100,
				},
			},
		})

		if got := m.statusLine(); got != "Syncing your account data." {
			t.Fatalf("statusLine() = %q", got)
		}
		if got := m.detailLine(); got != "25 / 100 rows" {
			t.Fatalf("detailLine() = %q", got)
		}
		pct, ok := m.percent()
		if !ok || pct != 25 {
			t.Fatalf("percent() = (%v, %t), want (25, true)", pct, ok)
		}
	})

	t.Run("ready and error states are projected", func(t *testing.T) {
		t.Parallel()

		m := New(theme.New(false))
		m.SetStatus(accountruntime.Status{Sync: &pssyncer.Ready{}})
		if got := m.statusLine(); got != "Your workspace is ready." {
			t.Fatalf("ready statusLine() = %q", got)
		}

		m.SetStatus(accountruntime.Status{Sync: &pssyncer.Error{Err: errors.New("boom")}})
		if got := m.statusLine(); got != "Sync failed: boom" {
			t.Fatalf("error statusLine() = %q", got)
		}
	})
}
