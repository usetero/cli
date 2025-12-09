package log_events_test

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/onboarding/log_events"
	"github.com/usetero/cli/internal/tui/onboarding/log_events/log_eventstest"
	"github.com/usetero/cli/internal/tui/tuitest"
)

func TestDiscoveryStep_Update(t *testing.T) {
	t.Run("completes when discovery status is ready", func(t *testing.T) {
		// Arrange
		percent := 100.0
		poller := &log_eventstest.MockLogDiscoveryProgressPoller{
			GetLogDiscoveryProgressFunc: func(ctx context.Context, datadogAccountID string) (*api.LogEventDiscoveryProgress, error) {
				return &api.LogEventDiscoveryProgress{
					Status:          api.DiscoveryStatusReady,
					PercentComplete: &percent,
					WeeklyVolume:    1000000,
				}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		ddAccountID := "dd-123"
		step := log_events.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, apiClient, logger, nil)

		// Act: run init command
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !updated.IsComplete() {
			t.Error("expected step to complete when discovery is ready")
		}
	})

	t.Run("continues polling when discovery in progress", func(t *testing.T) {
		// Arrange
		calls := 0
		percent := 50.0
		poller := &log_eventstest.MockLogDiscoveryProgressPoller{
			GetLogDiscoveryProgressFunc: func(ctx context.Context, datadogAccountID string) (*api.LogEventDiscoveryProgress, error) {
				calls++
				return &api.LogEventDiscoveryProgress{
					Status:          api.DiscoveryStatusDiscovering,
					PercentComplete: &percent,
					WeeklyVolume:    1000000,
				}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		ddAccountID := "dd-123"
		step := log_events.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, apiClient, logger, nil)

		// Act: run init command (first poll)
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert: not complete, poller was called
		if updated.IsComplete() {
			t.Error("expected step NOT to complete when discovery in progress")
		}
		if calls != 1 {
			t.Errorf("expected 1 poll call, got %d", calls)
		}
	})

	t.Run("sets error state on API failure", func(t *testing.T) {
		// Arrange
		poller := &log_eventstest.MockLogDiscoveryProgressPoller{
			GetLogDiscoveryProgressFunc: func(ctx context.Context, datadogAccountID string) (*api.LogEventDiscoveryProgress, error) {
				return nil, errors.New("API error")
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		ddAccountID := "dd-123"
		step := log_events.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, apiClient, logger, nil)

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

	t.Run("errors when no datadog account ID", func(t *testing.T) {
		// Arrange
		poller := &log_eventstest.MockLogDiscoveryProgressPoller{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := log_events.NewDiscoveryStep("admin", "org-1", "acc-1", nil, poller, apiClient, logger, nil)

		// Act: run init command
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !updated.HasError() {
			t.Error("expected error when no datadog account ID")
		}
	})
}
