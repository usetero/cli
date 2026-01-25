package datadog_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/datadog/datadogtest"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/tuitest"
)

// discoveryTestTheme creates a theme for testing
func discoveryTestTheme() *styles.Theme {
	return styles.NewTheme(true)
}

// isDiscoveryComplete checks if a step is complete by checking Next() returns nil error
func isDiscoveryComplete(s step.Step) bool {
	_, err := s.Next()
	return err == nil
}

func TestDiscoveryStep_Update(t *testing.T) {
	t.Parallel()
	testOrg := api.Organization{ID: "org-1", Name: "Test Org"}
	testAccount := api.Account{ID: "acc-1", Name: "Test Account"}

	t.Run("completes when status is ready", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:          api.DatadogAccountStatusReady,
					PercentComplete: 100,
					ServiceCount:    10,
					ReadyServices:   10,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !isDiscoveryComplete(updated) {
			t.Error("expected step to complete when status is ready")
		}
		if updated.HasError() {
			t.Errorf("expected no error, got: %v", updated.Error())
		}
	})

	t.Run("stays busy while status is analyzing", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:            api.DatadogAccountStatusAnalyzing,
					PercentComplete:   50,
					ServiceCount:      10,
					ReadyServices:     5,
					AnalyzingServices: 5,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if isDiscoveryComplete(updated) {
			t.Error("expected step to NOT complete while analyzing")
		}
		if !updated.IsBusy() {
			t.Error("expected step to be busy while analyzing")
		}
	})

	t.Run("sets error state on failure", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return nil, errors.New("API error")
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !updated.HasError() {
			t.Error("expected step to have error")
		}
		if isDiscoveryComplete(updated) {
			t.Error("expected step NOT to complete on error")
		}
	})

	t.Run("retries on enter key when error", func(t *testing.T) {
		t.Parallel()
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
					ServiceCount:    10,
					ReadyServices:   10,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// First attempt fails
		cmd := s.Init()
		updated := s
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
		if !isDiscoveryComplete(updated) {
			t.Error("expected step to complete after successful retry")
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("view shows volume when analyzing", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:            api.DatadogAccountStatusAnalyzing,
					PercentComplete:   67,
					ServiceLogVolume:  2400000,
					ServiceCount:      12,
					ActiveServices:    12,
					ReadyServices:     8,
					AnalyzingServices: 4,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert: check view contains expected content
		view := updated.View()
		if !contains(view, "2.4M logs") {
			t.Errorf("expected view to show '2.4M logs', got: %s", view)
		}
		if !contains(view, "12 services") {
			t.Errorf("expected view to show '12 services', got: %s", view)
		}
	})

	t.Run("shows inactive state when all services have zero volume", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:           api.DatadogAccountStatusInactive,
					ServiceCount:     19,
					InactiveServices: 19,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert: should not complete (user can't proceed)
		if isDiscoveryComplete(updated) {
			t.Error("expected step NOT to complete when all services are inactive")
		}

		// Assert: should stay busy (keeps polling)
		if !updated.IsBusy() {
			t.Error("expected step to stay busy when all services are inactive")
		}

		// Assert: should not show error (inactive is not an error)
		if updated.HasError() {
			t.Errorf("expected no error for inactive state, got: %v", updated.Error())
		}

		// Assert: view should show inactive messaging with yellow substatus
		view := updated.View()
		if !contains(view, "19 services") {
			t.Errorf("expected view to show '19 services', got: %s", view)
		}
		if !contains(view, "no recent log data") {
			t.Errorf("expected view to show 'no recent log data', got: %s", view)
		}
		if !contains(view, "Send logs to Datadog") {
			t.Errorf("expected view to show guidance 'Send logs to Datadog', got: %s", view)
		}
	})

	t.Run("does not show inactive state when some services are active", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:           api.DatadogAccountStatusAnalyzing,
					ServiceCount:     5,
					InactiveServices: 3,
					ReadyServices:    2,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert: view should NOT show inactive messaging (some services are active)
		view := updated.View()
		if contains(view, "no recent log data") {
			t.Errorf("expected view NOT to show inactive state when some services are active, got: %s", view)
		}
	})

	t.Run("shows error when status is DISABLED", func(t *testing.T) {
		t.Parallel()
		// Arrange - DISABLED during onboarding is a bug (auto-enable should have worked)
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:           api.DatadogAccountStatusDisabled,
					ServiceCount:     5,
					DisabledServices: 5,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act: run init command
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert: should show as error
		if !updated.HasError() {
			t.Error("expected HasError() to return true for DISABLED status")
		}

		// Assert: error message should mention disabled
		err := updated.Error()
		if err == nil || !contains(err.Error(), "disabled") {
			t.Errorf("expected error to mention 'disabled', got: %v", err)
		}
	})

	t.Run("shows stale state as our problem", func(t *testing.T) {
		t.Parallel()
		// Arrange - STALE means our system hasn't run in 48+ hours
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:        api.DatadogAccountStatusStale,
					ServiceCount:  19,
					StaleServices: 19,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert: view should indicate it's our problem
		view := updated.View()
		if !contains(view, "out of date") {
			t.Errorf("expected view to say 'out of date', got: %s", view)
		}
		if !contains(view, "48 hours") {
			t.Errorf("expected view to mention '48 hours', got: %s", view)
		}
		if !contains(view, "on our end") {
			t.Errorf("expected view to say 'on our end', got: %s", view)
		}

		// Should NOT tell user to send data (it's not their fault)
		if contains(view, "Send logs") {
			t.Errorf("should NOT tell user to send logs for STALE status, got: %s", view)
		}
	})

	t.Run("shows discovering state waiting for our analysis", func(t *testing.T) {
		t.Parallel()
		// Arrange - DISCOVERING with services but no progress means our pipeline is slow
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:              api.DatadogAccountStatusDiscovering,
					ServiceCount:        19,
					ActiveServices:      19,
					DiscoveringServices: 19,
					ReadyServices:       0,
					AnalyzingServices:   0,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert: view should say waiting for analysis (our system), not waiting for data (their fault)
		view := updated.View()
		if !contains(view, "19 services") {
			t.Errorf("expected view to show '19 services', got: %s", view)
		}
		if !contains(view, "waiting for analysis") {
			t.Errorf("expected view to say 'waiting for analysis', got: %s", view)
		}

		// Should NOT blame the user
		if contains(view, "Send logs") {
			t.Errorf("should NOT tell user to send logs when DISCOVERING, got: %s", view)
		}
	})

	t.Run("shows broken status as error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:         api.DatadogAccountStatusBroken,
					ServiceCount:   5,
					BrokenServices: 5,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !updated.HasError() {
			t.Error("expected HasError() to return true for BROKEN status")
		}
	})

	t.Run("shows broken services as issues when analyzing", func(t *testing.T) {
		t.Parallel()
		// Arrange - some services broken while others are fine
		ddAccountID := "dd-123"
		poller := &datadogtest.MockStatusPoller{
			GetStatusFunc: func(ctx context.Context, datadogAccountID string) (*api.DatadogAccountStatus, error) {
				return &api.DatadogAccountStatus{
					Status:            api.DatadogAccountStatusAnalyzing,
					ServiceCount:      10,
					ActiveServices:    10,
					ReadyServices:     5,
					AnalyzingServices: 3,
					BrokenServices:    2,
				}, nil
			},
		}
		logger := logtest.New(t)

		s := datadog.NewDiscoveryStep(discoveryTestTheme(), "admin", testOrg, testAccount, &ddAccountID, poller, logger, nil)

		// Act
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert: should surface broken services as an issue
		view := updated.View()
		if !contains(view, "2 services have errors") {
			t.Errorf("expected view to show broken services issue, got: %s", view)
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
