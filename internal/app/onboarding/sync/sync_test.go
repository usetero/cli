package sync

import (
	"testing"

	appmsg "github.com/usetero/cli/internal/app/msgs"
	onboardingmsg "github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/styles"
)

func TestInitEmitsSyncCompleteWhenReady(t *testing.T) {
	t.Parallel()

	syncer := powersynctest.NewMockSyncer()
	syncer.IsReadyFunc = func() bool { return true }
	m := New(styles.NewTheme(true), syncer, logtest.NewScope(t))

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected non-nil init cmd")
	}
	msg := cmd()
	if _, ok := msg.(onboardingmsg.SyncComplete); !ok {
		t.Fatalf("expected SyncComplete message, got %T", msg)
	}
}

func TestUpdateEmitsSyncCompleteOnReadyState(t *testing.T) {
	t.Parallel()

	syncer := powersynctest.NewMockSyncer()
	m := New(styles.NewTheme(true), syncer, logtest.NewScope(t))

	cmd := m.Update(appmsg.SyncStateChanged{State: powersync.NewReady()})
	if cmd == nil {
		t.Fatal("expected non-nil command on ready state")
	}
	msg := cmd()
	if _, ok := msg.(onboardingmsg.SyncComplete); !ok {
		t.Fatalf("expected SyncComplete message, got %T", msg)
	}
}

func TestUpdateIgnoresNonReadySyncState(t *testing.T) {
	t.Parallel()

	syncer := powersynctest.NewMockSyncer()
	m := New(styles.NewTheme(true), syncer, logtest.NewScope(t))

	cmd := m.Update(appmsg.SyncStateChanged{State: powersync.NewConnecting()})
	if cmd != nil {
		t.Fatal("expected nil command for non-ready sync state")
	}
}
