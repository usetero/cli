package datadog_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/datadog/datadogtest"
	"github.com/usetero/cli/internal/tui/tuitest"
)

func TestDiscoveryStep_Update(t *testing.T) {
	t.Run("completes when status is ready", func(t *testing.T) {
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:          api.DatadogAccountStatusReady,
					PercentComplete: 100,
					Total:           10,
					Ready:           10,
				}, nil
			},
		}
		logger := logtest.New(t)

		step := datadog.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !updated.IsComplete() {
			t.Error("expected step to complete when status is ready")
		}
		if updated.HasError() {
			t.Errorf("expected no error, got: %v", updated.Error())
		}
	})

	t.Run("stays busy while status is in progress", func(t *testing.T) {
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:          api.DatadogAccountStatusInProgress,
					PercentComplete: 50,
					Total:           10,
					Ready:           5,
				}, nil
			},
		}
		logger := logtest.New(t)

		step := datadog.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if updated.IsComplete() {
			t.Error("expected step to NOT complete while in progress")
		}
		if !updated.IsBusy() {
			t.Error("expected step to be busy while in progress")
		}
	})

	t.Run("sets error state on failure", func(t *testing.T) {
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return nil, errors.New("API error")
			},
		}
		logger := logtest.New(t)

		step := datadog.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !updated.HasError() {
			t.Error("expected step to have error")
		}
		if updated.IsComplete() {
			t.Error("expected step NOT to complete on error")
		}
	})

	t.Run("retries on enter key when error", func(t *testing.T) {
		// Arrange
		ddAccountID := "dd-123"
		attempts := 0
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				attempts++
				if attempts == 1 {
					return nil, errors.New("first attempt fails")
				}
				return &api.DatadogAccountStatus{
					Status:          api.DatadogAccountStatusReady,
					PercentComplete: 100,
					Total:           10,
					Ready:           10,
				}, nil
			},
		}
		logger := logtest.New(t)

		step := datadog.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, logger, nil)

		// First attempt fails
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		if !updated.HasError() {
			t.Fatal("expected error after first attempt")
		}

		// Act: press enter to retry
		updated, cmd = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if updated.HasError() {
			t.Errorf("expected error to be cleared after retry, got: %v", updated.Error())
		}
		if !updated.IsComplete() {
			t.Error("expected step to complete after successful retry")
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("view shows progress percentage correctly", func(t *testing.T) {
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:          api.DatadogAccountStatusInProgress,
					PercentComplete: 97,
					Total:           19,
					Ready:           18,
				}, nil
			},
		}
		logger := logtest.New(t)

		step := datadog.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert: check view contains expected content
		view := updated.View()
		if !contains(view, "18 of 19") {
			t.Errorf("expected view to show '18 of 19', got: %s", view)
		}
		// The progress bar should show ~97%, not 1%
		// We check that it doesn't show a very low percentage
		if contains(view, "1%") && !contains(view, "91%") && !contains(view, "100%") {
			t.Errorf("expected view to show high percentage (97%%), but got low percentage")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
