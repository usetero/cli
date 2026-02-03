package sync

import (
	"errors"
	"fmt"
	"testing"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

func TestModel_Init(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when already ready", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			IsReadyFunc: func() bool { return true },
		}
		m := newTestModel(t, syncer)

		cmd := m.Init()

		if cmd != nil {
			t.Error("expected nil cmd when already ready")
		}
	})

	t.Run("starts polling when not ready", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			IsReadyFunc: func() bool { return false },
		}
		m := newTestModel(t, syncer)

		cmd := m.Init()

		if cmd == nil {
			t.Error("expected cmd to start polling")
		}
	})
}

func TestModel_Update(t *testing.T) {
	t.Parallel()

	t.Run("continues polling while not ready", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			IsReadyFunc: func() bool { return false },
			StateFunc:   func() powersync.State { return powersync.NewSyncing("Syncing...") },
		}
		m := newTestModel(t, syncer)

		updated, cmd := m.Update(pollMsg{})

		// cmd is a tea.Tick - just check it's not nil (don't execute, it blocks)
		if cmd == nil {
			t.Error("expected cmd to continue polling")
		}
		if !updated.IsBusy() {
			t.Error("expected IsBusy to be true while syncing")
		}
	})

	t.Run("emits ready message when sync completes", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			IsReadyFunc: func() bool { return true },
		}
		m := newTestModel(t, syncer)

		_, cmd := m.Update(pollMsg{})

		if cmd == nil {
			t.Fatal("expected cmd when sync becomes ready")
		}

		// This cmd returns readyMsg immediately (not a tick)
		msg := cmd()
		if _, ok := msg.(readyMsg); !ok {
			t.Errorf("expected readyMsg, got %T", msg)
		}
	})
}

func TestModel_Next(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotReady while syncing", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			IsReadyFunc: func() bool { return false },
		}
		m := newTestModel(t, syncer)

		_, err := m.Next()

		if !errors.Is(err, step.ErrNotReady) {
			t.Errorf("expected ErrNotReady, got %v", err)
		}
	})

	t.Run("returns nil step when ready (flow complete)", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			IsReadyFunc: func() bool { return true },
		}
		m := newTestModel(t, syncer)

		nextStep, err := m.Next()

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if nextStep != nil {
			t.Error("expected nil step (flow complete)")
		}
	})
}

func TestModel_IsBusy(t *testing.T) {
	t.Parallel()

	t.Run("returns true while syncing", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			IsReadyFunc: func() bool { return false },
		}
		m := newTestModel(t, syncer)

		if !m.IsBusy() {
			t.Error("expected IsBusy to be true")
		}
	})

	t.Run("returns false when ready", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			IsReadyFunc: func() bool { return true },
		}
		m := newTestModel(t, syncer)

		if m.IsBusy() {
			t.Error("expected IsBusy to be false")
		}
	})
}

func TestModel_HasError(t *testing.T) {
	t.Parallel()

	t.Run("returns false when no error", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			StateFunc: func() powersync.State { return powersync.NewDisconnected() },
		}
		m := newTestModel(t, syncer)

		if m.HasError() {
			t.Error("expected HasError to be false")
		}
	})

	t.Run("returns true when syncer has error", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			StateFunc: func() powersync.State {
				return powersync.NewError(fmt.Errorf("sync failed"))
			},
		}
		m := newTestModel(t, syncer)

		if !m.HasError() {
			t.Error("expected HasError to be true")
		}
	})
}

func TestModel_Error(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when no error", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			StateFunc: func() powersync.State { return powersync.NewDisconnected() },
		}
		m := newTestModel(t, syncer)

		if m.Error() != nil {
			t.Error("expected Error to be nil")
		}
	})

	t.Run("returns syncer error", func(t *testing.T) {
		t.Parallel()

		expectedErr := fmt.Errorf("sync failed")
		syncer := &powersynctest.MockSyncer{
			StateFunc: func() powersync.State {
				return powersync.NewError(expectedErr)
			},
		}
		m := newTestModel(t, syncer)

		if !errors.Is(m.Error(), expectedErr) {
			t.Errorf("expected Error to be %v, got %v", expectedErr, m.Error())
		}
	})
}

func TestModel_View(t *testing.T) {
	t.Parallel()

	t.Run("shows progress while syncing", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			IsReadyFunc: func() bool { return false },
			StateFunc:   func() powersync.State { return powersync.NewSyncing("Syncing...") },
		}
		m := newTestModel(t, syncer)

		view := m.View()

		if view == "" {
			t.Error("expected non-empty view")
		}
	})

	t.Run("shows ready when complete", func(t *testing.T) {
		t.Parallel()

		syncer := &powersynctest.MockSyncer{
			IsReadyFunc: func() bool { return true },
		}
		m := newTestModel(t, syncer)

		view := m.View()

		if view == "" {
			t.Error("expected non-empty view")
		}
	})
}

func newTestModel(t *testing.T, syncer *powersynctest.MockSyncer) Model {
	theme := styles.NewTheme(true)
	logger := logtest.New(t)

	return New(
		theme,
		domain.Organization{ID: "org-1"},
		domain.Account{ID: "acc-1"},
		domain.Workspace{ID: "ws-1"},
		syncer,
		logger,
	)
}
