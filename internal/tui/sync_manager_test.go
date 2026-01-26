package tui_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/tui"
	"github.com/usetero/cli/internal/tui/tuitest"
)

// mockSyncer implements tui.Syncer for testing.
type mockSyncer struct {
	startFunc            func(ctx context.Context, db sqlite.Database, accountID, token string) error
	stopFunc             func()
	statusFunc           func() powersync.Status
	waitForFirstSyncFunc func(ctx context.Context) error

	// Track calls
	startCalled bool
	stopCalled  bool
	db          sqlite.Database
	accountID   string
	token       string
}

func (m *mockSyncer) Start(ctx context.Context, db sqlite.Database, accountID, token string) error {
	m.startCalled = true
	m.db = db
	m.accountID = accountID
	m.token = token
	if m.startFunc != nil {
		return m.startFunc(ctx, db, accountID, token)
	}
	return nil
}

func (m *mockSyncer) Stop() {
	m.stopCalled = true
	if m.stopFunc != nil {
		m.stopFunc()
	}
}

func (m *mockSyncer) Status() powersync.Status {
	if m.statusFunc != nil {
		return m.statusFunc()
	}
	return powersync.StatusDisconnected
}

func (m *mockSyncer) WaitForFirstSync(ctx context.Context) error {
	if m.waitForFirstSyncFunc != nil {
		return m.waitForFirstSyncFunc(ctx)
	}
	return nil
}

// mockTokenRefresher implements powersync.TokenRefresher for testing.
type mockTokenRefresher struct {
	token string
	err   error
}

func (m *mockTokenRefresher) GetAccessToken(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

func TestSyncManager_Update(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for unhandled messages", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		sm := tui.NewSyncManagerForTest(
			context.Background(),
			&powersync.Config{},
			&mockSyncer{},
			&mockTokenRefresher{token: "token"},
			logger,
		)

		cmd := sm.Update(tea.KeyPressMsg{})
		if cmd != nil {
			t.Error("expected nil command for unhandled message")
		}
	})

	t.Run("starts sync command when account is selected", func(t *testing.T) {
		t.Parallel()

		syncer := &mockSyncer{
			statusFunc: func() powersync.Status {
				return powersync.StatusConnecting
			},
		}
		logger := logtest.New(t)

		sm := tui.NewSyncManagerForTest(
			context.Background(),
			&powersync.Config{},
			syncer,
			&mockTokenRefresher{token: "test-token"},
			logger,
		)

		// Send AccountSelectedMsg - should return a command
		msg := tui.AccountSelectedMsg{
			Account: api.Account{ID: "acc-123", Name: "Test Account"},
		}
		cmd := sm.Update(msg)

		if cmd == nil {
			t.Error("expected command to be returned")
		}
	})

	t.Run("emits InitialSyncCompletedMsg after successful sync chain", func(t *testing.T) {
		t.Parallel()

		syncer := &mockSyncer{
			statusFunc: func() powersync.Status {
				return powersync.StatusConnected
			},
		}
		logger := logtest.New(t)

		sm := tui.NewSyncManagerForTest(
			context.Background(),
			&powersync.Config{},
			syncer,
			&mockTokenRefresher{token: "test-token"},
			logger,
		)

		// Simulate the message chain that happens after sync starts successfully
		// 1. syncStartedMsg with a mock database
		db := sqlitetest.NewMockDB()
		msgs := tuitest.DrainCmds(sm.Update(tui.SyncStartedMsgForTest(db, nil)))

		// Should get a command to wait for initial sync
		if len(msgs) == 0 {
			t.Fatal("expected command after syncStartedMsg")
		}

		// 2. Execute those commands and look for InitialSyncCompletedMsg
		foundInitialSyncCompleted := false
		for _, m := range msgs {
			cmd := sm.Update(m)
			for _, m2 := range tuitest.DrainCmds(cmd) {
				if _, ok := m2.(tui.InitialSyncCompletedMsg); ok {
					foundInitialSyncCompleted = true
				}
			}
		}

		if !foundInitialSyncCompleted {
			t.Error("expected InitialSyncCompletedMsg to be emitted")
		}
	})

	t.Run("does not emit InitialSyncCompletedMsg on sync error", func(t *testing.T) {
		t.Parallel()

		syncer := &mockSyncer{
			waitForFirstSyncFunc: func(ctx context.Context) error {
				return errors.New("sync failed")
			},
		}
		logger := logtest.New(t)

		sm := tui.NewSyncManagerForTest(
			context.Background(),
			&powersync.Config{},
			syncer,
			&mockTokenRefresher{token: "test-token"},
			logger,
		)

		// Simulate syncStartedMsg
		db := sqlitetest.NewMockDB()
		msgs := tuitest.DrainCmds(sm.Update(tui.SyncStartedMsgForTest(db, nil)))

		// Execute commands - should NOT emit InitialSyncCompletedMsg
		for _, m := range msgs {
			cmd := sm.Update(m)
			for _, m2 := range tuitest.DrainCmds(cmd) {
				if _, ok := m2.(tui.InitialSyncCompletedMsg); ok {
					t.Error("should not emit InitialSyncCompletedMsg on error")
				}
			}
		}
	})
}

func TestSyncManager_Shutdown(t *testing.T) {
	t.Parallel()

	t.Run("stops sync and closes database", func(t *testing.T) {
		t.Parallel()

		syncer := &mockSyncer{}
		logger := logtest.New(t)

		sm := tui.NewSyncManagerForTest(
			context.Background(),
			&powersync.Config{},
			syncer,
			&mockTokenRefresher{token: "test-token"},
			logger,
		)

		// Simulate sync started with a database
		db := sqlitetest.NewMockDB()
		sm.Update(tui.SyncStartedMsgForTest(db, nil))

		// Shutdown
		sm.Shutdown()

		if !syncer.stopCalled {
			t.Error("expected sync.Stop to be called")
		}
		if !db.Closed {
			t.Error("expected database to be closed")
		}
	})
}
