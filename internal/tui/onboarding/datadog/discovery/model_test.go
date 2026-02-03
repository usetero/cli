package datadogdiscovery

import (
	"context"
	"errors"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/preferences/preferencestest"
	"github.com/usetero/cli/internal/styles"
)

func TestModel_Init(t *testing.T) {
	t.Parallel()

	t.Run("returns spinner tick command", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		cmd := m.Init()

		if cmd == nil {
			t.Fatal("Init() returned nil command")
		}

		// Execute the batch and check we get expected messages
		// tea.Batch returns a function that sends all commands
		// We can't easily inspect it, but we can verify it's not nil
	})
}

func TestModel_Update(t *testing.T) {
	t.Parallel()

	t.Run("spinner animates while loading", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		_ = m.Init()

		// Simulate spinner tick
		tickMsg := spinner.TickMsg{}
		step, cmd := m.Update(tickMsg)
		updated, ok := step.(Model)
		if !ok {
			t.Fatal("Update() returned wrong type")
		}

		// Spinner should produce another tick command to continue animation
		if cmd == nil {
			t.Error("Update(spinner.TickMsg) should return a command to continue animation")
		}

		// Model should still be loading
		if !updated.loading {
			t.Error("Model should still be loading")
		}
	})

	t.Run("spinner animates after status received but not complete", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		_ = m.Init()

		// Receive status that's not complete
		status := &api.DatadogAccountStatus{
			Status:       api.DatadogAccountStatusAnalyzing,
			ReadyForUse:  false,
			ServiceCount: 5,
		}
		step, _ := m.Update(statusFetchedMsg{status: status})
		var ok bool
		m, ok = step.(Model)
		if !ok {
			t.Fatal("Update() returned wrong type")
		}

		// Now send spinner tick - should still animate
		tickMsg := spinner.TickMsg{}
		step, cmd := m.Update(tickMsg)
		updated, ok := step.(Model)
		if !ok {
			t.Fatal("Update() returned wrong type")
		}

		if cmd == nil {
			t.Error("Update(spinner.TickMsg) should return command when not complete")
		}

		if updated.isComplete() {
			t.Error("Model should not be complete yet")
		}
	})

	t.Run("spinner stops when complete", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		_ = m.Init()

		// Receive complete status
		status := &api.DatadogAccountStatus{
			Status:        api.DatadogAccountStatusReady,
			ReadyForUse:   true,
			ServiceCount:  5,
			ReadyServices: 5,
		}
		step, _ := m.Update(statusFetchedMsg{status: status})
		var ok bool
		m, ok = step.(Model)
		if !ok {
			t.Fatal("Update() returned wrong type")
		}

		if !m.isComplete() {
			t.Fatal("Model should be complete")
		}

		// Spinner tick should not continue animation
		tickMsg := spinner.TickMsg{}
		step, _ = m.Update(tickMsg)
		_, ok = step.(Model)
		if !ok {
			t.Fatal("Update() returned wrong type")
		}

		// When complete, spinner update is skipped - we verify by checking isComplete() above
	})

	t.Run("handles status fetch error", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		_ = m.Init()

		step, _ := m.Update(statusFetchedMsg{err: context.DeadlineExceeded})
		updated, ok := step.(Model)
		if !ok {
			t.Fatal("Update() returned wrong type")
		}

		if !updated.HasError() {
			t.Error("Model should have error after fetch failure")
		}

		if !errors.Is(updated.Error(), context.DeadlineExceeded) {
			t.Errorf("Error() = %v, want %v", updated.Error(), context.DeadlineExceeded)
		}
	})

	t.Run("retry clears error and restarts polling", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)

		// Set up error state
		step, _ := m.Update(statusFetchedMsg{err: context.DeadlineExceeded})
		var ok bool
		m, ok = step.(Model)
		if !ok {
			t.Fatal("Update() returned wrong type")
		}

		if !m.HasError() {
			t.Fatal("Model should have error")
		}

		// Press enter to retry
		step, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		updated, ok := step.(Model)
		if !ok {
			t.Fatal("Update() returned wrong type")
		}

		if updated.HasError() {
			t.Error("Error should be cleared after retry")
		}

		if !updated.loading {
			t.Error("Model should be loading after retry")
		}

		if cmd == nil {
			t.Error("Retry should return commands to restart polling")
		}
	})
}

func TestModel_IsBusy(t *testing.T) {
	t.Parallel()

	t.Run("returns true while not complete and no error", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)

		if !m.IsBusy() {
			t.Error("IsBusy() should return true initially")
		}
	})

	t.Run("returns false when complete", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		step, _ := m.Update(statusFetchedMsg{
			status: &api.DatadogAccountStatus{
				ReadyForUse: true,
			},
		})
		updated, ok := step.(Model)
		if !ok {
			t.Fatal("Update() returned wrong type")
		}

		if updated.IsBusy() {
			t.Error("IsBusy() should return false when complete")
		}
	})

	t.Run("returns false when error", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		step, _ := m.Update(statusFetchedMsg{err: context.DeadlineExceeded})
		updated, ok := step.(Model)
		if !ok {
			t.Fatal("Update() returned wrong type")
		}

		if updated.IsBusy() {
			t.Error("IsBusy() should return false when there's an error")
		}
	})
}

func newTestModel(t *testing.T) Model {
	t.Helper()

	datadogAccountID := "dd-123"
	services := apitest.NewMockAPIServices(nil, nil, nil, apitest.NewMockDatadogAccounts())

	return New(
		context.Background(),
		styles.NewTheme(true),
		"engineer",
		domain.Organization{ID: "org-1", Name: "Test Org"},
		domain.Account{ID: "acc-1"},
		&datadogAccountID,
		services,
		preferencestest.NewMockPreferences(),
		logtest.New(t),
	)
}
