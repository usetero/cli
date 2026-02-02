package datadog_test

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/preferences/preferencestest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/tuitest"
)

// apiKeyTestTheme creates a theme for testing
func apiKeyTestTheme() *styles.Theme {
	return styles.NewTheme(true)
}

// isAPIKeyComplete checks if a step is complete by checking Next() returns nil error
func isAPIKeyComplete(s step.Step) bool {
	_, err := s.Next()
	return err == nil
}

func TestAPIKeyStep_Update(t *testing.T) {
	t.Parallel()
	org := apitest.NewOrganization()
	account := apitest.NewAccount()

	t.Run("validates API key on enter", func(t *testing.T) {
		t.Parallel()
		// Arrange
		validated := false
		ddAccounts := &apitest.MockDatadogAccounts{
			ValidateAPIKeyFunc: func(ctx context.Context, apiKey string, site string) (bool, string, error) {
				validated = true
				return true, "", nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		prefs := &preferencestest.MockPreferences{}
		s := datadog.NewAPIKeyStep(context.Background(), apiKeyTestTheme(), "admin", org, account, "US1", ddAccounts, workspaces, prefs, apiClient, logger, nil)

		// Transition to input screen (press enter on interstitial)
		updated, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Type API key
		for _, r := range "test-api-key-123" {
			updated, _ = updated.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}

		// Act: press enter to submit
		updated, cmd = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !validated {
			t.Error("expected API key to be validated")
		}
		if !isAPIKeyComplete(updated) {
			t.Error("expected step to complete after validation")
		}
	})

	t.Run("sets error state on invalid key", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccounts := &apitest.MockDatadogAccounts{
			ValidateAPIKeyFunc: func(ctx context.Context, apiKey string, site string) (bool, string, error) {
				return false, "Invalid API key", nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		prefs := &preferencestest.MockPreferences{}
		s := datadog.NewAPIKeyStep(context.Background(), apiKeyTestTheme(), "admin", org, account, "US1", ddAccounts, workspaces, prefs, apiClient, logger, nil)

		// Transition to input screen
		updated, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Type API key
		for _, r := range "bad-key" {
			updated, _ = updated.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}

		// Submit
		updated, cmd = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !updated.HasError() {
			t.Error("expected step to have error")
		}
		if isAPIKeyComplete(updated) {
			t.Error("expected step NOT to complete on error")
		}
	})

	t.Run("does not submit empty input", func(t *testing.T) {
		t.Parallel()
		// Arrange
		validated := false
		ddAccounts := &apitest.MockDatadogAccounts{
			ValidateAPIKeyFunc: func(ctx context.Context, apiKey string, site string) (bool, string, error) {
				validated = true
				return true, "", nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		prefs := &preferencestest.MockPreferences{}
		s := datadog.NewAPIKeyStep(context.Background(), apiKeyTestTheme(), "admin", org, account, "US1", ddAccounts, workspaces, prefs, apiClient, logger, nil)

		// Transition to input screen
		updated, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Act: press enter without typing anything
		updated, cmd = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if validated {
			t.Error("expected validation NOT to be called with empty input")
		}
		if isAPIKeyComplete(updated) {
			t.Error("expected step NOT to complete with empty input")
		}
	})

	t.Run("transitions from interstitial to input on enter", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccounts := &apitest.MockDatadogAccounts{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		prefs := &preferencestest.MockPreferences{}
		s := datadog.NewAPIKeyStep(context.Background(), apiKeyTestTheme(), "admin", org, account, "US1", ddAccounts, workspaces, prefs, apiClient, logger, nil)

		// Assert: not complete initially
		if isAPIKeyComplete(s) {
			t.Error("expected step NOT to be complete initially")
		}

		// Act: press enter to transition
		updated, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Assert: still not complete (just transitioned to input)
		if isAPIKeyComplete(updated) {
			t.Error("expected step NOT to be complete after transition")
		}
	})
}
