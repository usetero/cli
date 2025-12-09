package services_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/onboarding/services"
	"github.com/usetero/cli/internal/tui/onboarding/services/servicestest"
	"github.com/usetero/cli/internal/tui/tuitest"
)

func TestDiscoveryStep_Update(t *testing.T) {
	t.Run("completes when discovery status is ready", func(t *testing.T) {
		// Arrange
		poller := &servicestest.MockServiceDiscoveryPoller{
			GetServiceDiscoveryStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.ServiceDiscoveryStatus, error) {
				return &api.ServiceDiscoveryStatus{
					Status:             api.DiscoveryStatusReady,
					ServicesDiscovered: 5,
				}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		ddAccountID := "dd-123"
		step := services.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, apiClient, logger, nil)

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
		poller := &servicestest.MockServiceDiscoveryPoller{
			GetServiceDiscoveryStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.ServiceDiscoveryStatus, error) {
				calls++
				return &api.ServiceDiscoveryStatus{
					Status:             api.DiscoveryStatusDiscovering,
					ServicesDiscovered: calls,
				}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		ddAccountID := "dd-123"
		step := services.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, apiClient, logger, nil)

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
		poller := &servicestest.MockServiceDiscoveryPoller{
			GetServiceDiscoveryStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.ServiceDiscoveryStatus, error) {
				return nil, errors.New("API error")
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		ddAccountID := "dd-123"
		step := services.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, apiClient, logger, nil)

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

	t.Run("retries on enter when error", func(t *testing.T) {
		// Arrange
		attempts := 0
		poller := &servicestest.MockServiceDiscoveryPoller{
			GetServiceDiscoveryStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.ServiceDiscoveryStatus, error) {
				attempts++
				if attempts == 1 {
					return nil, errors.New("first attempt fails")
				}
				return &api.ServiceDiscoveryStatus{
					Status:             api.DiscoveryStatusReady,
					ServicesDiscovered: 3,
				}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		ddAccountID := "dd-123"
		step := services.NewDiscoveryStep("admin", "org-1", "acc-1", &ddAccountID, poller, apiClient, logger, nil)

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
			t.Error("expected error to be cleared after retry")
		}
		if !updated.IsComplete() {
			t.Error("expected step to complete after successful retry")
		}
	})

	t.Run("errors when no datadog account ID", func(t *testing.T) {
		// Arrange
		poller := &servicestest.MockServiceDiscoveryPoller{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := services.NewDiscoveryStep("admin", "org-1", "acc-1", nil, poller, apiClient, logger, nil)

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
