package datadog_test

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/datadog/datadogtest"
	"github.com/usetero/cli/internal/tui/tuitest"
)

func TestAPIKeyStep_Update(t *testing.T) {
	t.Run("validates API key on enter", func(t *testing.T) {
		// Arrange
		validated := false
		validator := &datadogtest.MockAPIKeyValidator{
			ValidateAPIKeyFunc: func(ctx context.Context, apiKey string, site string) (bool, string, error) {
				validated = true
				return true, "", nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := datadog.NewAPIKeyStep("admin", "org-1", "acc-1", "US1", validator, apiClient, logger, nil)

		// Transition to input screen (press enter on interstitial)
		updated, cmd := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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
		if !updated.IsComplete() {
			t.Error("expected step to complete after validation")
		}
	})

	t.Run("sets error state on invalid key", func(t *testing.T) {
		// Arrange
		validator := &datadogtest.MockAPIKeyValidator{
			ValidateAPIKeyFunc: func(ctx context.Context, apiKey string, site string) (bool, string, error) {
				return false, "Invalid API key", nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := datadog.NewAPIKeyStep("admin", "org-1", "acc-1", "US1", validator, apiClient, logger, nil)

		// Transition to input screen
		updated, cmd := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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
		if updated.IsComplete() {
			t.Error("expected step NOT to complete on error")
		}
	})

	t.Run("does not submit empty input", func(t *testing.T) {
		// Arrange
		validated := false
		validator := &datadogtest.MockAPIKeyValidator{
			ValidateAPIKeyFunc: func(ctx context.Context, apiKey string, site string) (bool, string, error) {
				validated = true
				return true, "", nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := datadog.NewAPIKeyStep("admin", "org-1", "acc-1", "US1", validator, apiClient, logger, nil)

		// Transition to input screen
		updated, cmd := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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
		if updated.IsComplete() {
			t.Error("expected step NOT to complete with empty input")
		}
	})

	t.Run("transitions from interstitial to input on enter", func(t *testing.T) {
		// Arrange
		validator := &datadogtest.MockAPIKeyValidator{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := datadog.NewAPIKeyStep("admin", "org-1", "acc-1", "US1", validator, apiClient, logger, nil)

		// Assert: not complete initially
		if step.IsComplete() {
			t.Error("expected step NOT to be complete initially")
		}

		// Act: press enter to transition
		updated, _ := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Assert: still not complete (just transitioned to input)
		if updated.IsComplete() {
			t.Error("expected step NOT to be complete after transition")
		}
	})
}
