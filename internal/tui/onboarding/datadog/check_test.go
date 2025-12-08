package datadog_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/datadog/datadogtest"
	"github.com/usetero/cli/internal/tui/tuitest"
)

func keyMsg(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func TestCheckDatadogStep_Update(t *testing.T) {
	t.Run("completes when datadog account exists", func(t *testing.T) {
		// Arrange
		checker := &datadogtest.MockDatadogAccountChecker{
			HasAccountFunc: func(ctx context.Context, accountID string) (bool, error) {
				return true, nil
			},
			GetAccountFunc: func(ctx context.Context, accountID string) (*api.DatadogAccount, error) {
				return &api.DatadogAccount{ID: "dd-1", Name: "Production DD"}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := datadog.NewCheckDatadogStep("admin", "org-1", "acc-1", checker, apiClient, logger, nil)

		// Act: run init command
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Assert
		if !updated.IsComplete() {
			t.Error("expected step to complete when datadog exists")
		}
		checkStep := updated.(*datadog.CheckDatadogStep)
		if checkStep.NeedsDatadogSetup() {
			t.Error("expected NeedsDatadogSetup to be false when datadog exists")
		}
	})

	t.Run("completes when no datadog account", func(t *testing.T) {
		// Arrange
		checker := &datadogtest.MockDatadogAccountChecker{
			HasAccountFunc: func(ctx context.Context, accountID string) (bool, error) {
				return false, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := datadog.NewCheckDatadogStep("admin", "org-1", "acc-1", checker, apiClient, logger, nil)

		// Act: run init command
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Assert
		if !updated.IsComplete() {
			t.Error("expected step to complete when no datadog")
		}
		checkStep := updated.(*datadog.CheckDatadogStep)
		if !checkStep.NeedsDatadogSetup() {
			t.Error("expected NeedsDatadogSetup to be true when no datadog")
		}
	})

	t.Run("sets error state on failure", func(t *testing.T) {
		// Arrange
		checker := &datadogtest.MockDatadogAccountChecker{
			HasAccountFunc: func(ctx context.Context, accountID string) (bool, error) {
				return false, errors.New("API error")
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := datadog.NewCheckDatadogStep("admin", "org-1", "acc-1", checker, apiClient, logger, nil)

		// Act: run init command
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Assert
		if !updated.HasError() {
			t.Error("expected step to have error")
		}
		if updated.IsComplete() {
			t.Error("expected step NOT to complete on error")
		}
	})

	t.Run("retries on r key when error", func(t *testing.T) {
		// Arrange
		attempts := 0
		checker := &datadogtest.MockDatadogAccountChecker{
			HasAccountFunc: func(ctx context.Context, accountID string) (bool, error) {
				attempts++
				if attempts == 1 {
					return false, errors.New("first attempt fails")
				}
				return true, nil
			},
			GetAccountFunc: func(ctx context.Context, accountID string) (*api.DatadogAccount, error) {
				return &api.DatadogAccount{ID: "dd-1"}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := datadog.NewCheckDatadogStep("admin", "org-1", "acc-1", checker, apiClient, logger, nil)

		// First attempt fails
		cmd := step.Init()
		updated := step
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		if !updated.HasError() {
			t.Fatal("expected error after first attempt")
		}

		// Act: press 'r' to retry
		updated, cmd = updated.Update(keyMsg('r'))
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Assert
		if updated.HasError() {
			t.Error("expected error to be cleared after retry")
		}
		if !updated.IsComplete() {
			t.Error("expected step to complete after successful retry")
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})
}
