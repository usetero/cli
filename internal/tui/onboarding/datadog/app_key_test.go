package datadog_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/tuitest"
)

// appKeyTestTheme creates a theme for testing
func appKeyTestTheme() *styles.Theme {
	return styles.NewTheme(true)
}

// isAppKeyComplete checks if a step is complete by checking Next() returns nil error
func isAppKeyComplete(s step.Step) bool {
	_, err := s.Next()
	return err == nil
}

func TestAppKeyStep_Update(t *testing.T) {
	t.Parallel()
	testOrg := api.Organization{ID: "org-1", Name: "Test Org"}
	testAccount := api.Account{ID: "acc-1", Name: "Test Account"}

	t.Run("creates datadog account on enter", func(t *testing.T) {
		t.Parallel()
		// Arrange
		created := false
		ddAccounts := &apitest.MockDatadogAccounts{
			CreateAccountFunc: func(ctx context.Context, accountID string, name string, site string, apiKey string, appKey string) (*api.DatadogAccount, error) {
				created = true
				return &api.DatadogAccount{ID: "dd-1", Site: site}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := datadog.NewAppKeyStep(context.Background(), appKeyTestTheme(), "admin", testOrg, testAccount, "US1", "api-key-123", ddAccounts, apiClient, logger, nil)

		// Transition to input screen
		updated, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Type app key
		for _, r := range "app-key-123" {
			updated, _ = updated.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}

		// Act: press enter to submit
		updated, cmd = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !created {
			t.Error("expected datadog account to be created")
		}
		if !isAppKeyComplete(updated) {
			t.Error("expected step to complete after creation")
		}
	})

	t.Run("sets error state on failure", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccounts := &apitest.MockDatadogAccounts{
			CreateAccountFunc: func(ctx context.Context, accountID string, name string, site string, apiKey string, appKey string) (*api.DatadogAccount, error) {
				return nil, errors.New("invalid application key")
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := datadog.NewAppKeyStep(context.Background(), appKeyTestTheme(), "admin", testOrg, testAccount, "US1", "api-key-123", ddAccounts, apiClient, logger, nil)

		// Transition to input screen
		updated, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Type app key
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
		if isAppKeyComplete(updated) {
			t.Error("expected step NOT to complete on error")
		}
	})

	t.Run("does not submit empty input", func(t *testing.T) {
		t.Parallel()
		// Arrange
		created := false
		ddAccounts := &apitest.MockDatadogAccounts{
			CreateAccountFunc: func(ctx context.Context, accountID string, name string, site string, apiKey string, appKey string) (*api.DatadogAccount, error) {
				created = true
				return &api.DatadogAccount{ID: "dd-1"}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := datadog.NewAppKeyStep(context.Background(), appKeyTestTheme(), "admin", testOrg, testAccount, "US1", "api-key-123", ddAccounts, apiClient, logger, nil)

		// Transition to input screen
		updated, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Act: press enter without typing
		updated, cmd = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if created {
			t.Error("expected account NOT to be created with empty input")
		}
		if isAppKeyComplete(updated) {
			t.Error("expected step NOT to complete with empty input")
		}
	})

	t.Run("retries on enter when error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		attempts := 0
		ddAccounts := &apitest.MockDatadogAccounts{
			CreateAccountFunc: func(ctx context.Context, accountID string, name string, site string, apiKey string, appKey string) (*api.DatadogAccount, error) {
				attempts++
				if attempts == 1 {
					return nil, errors.New("first attempt fails")
				}
				return &api.DatadogAccount{ID: "dd-1", Site: site}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := datadog.NewAppKeyStep(context.Background(), appKeyTestTheme(), "admin", testOrg, testAccount, "US1", "api-key-123", ddAccounts, apiClient, logger, nil)

		// Transition to input screen
		updated, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Type app key and submit (first attempt fails)
		for _, r := range "app-key" {
			updated, _ = updated.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		updated, cmd = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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
		if !isAppKeyComplete(updated) {
			t.Error("expected step to complete after successful retry")
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})
}
