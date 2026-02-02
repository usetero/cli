package datadog_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/preferences/preferencestest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/tuitest"
)

func checkOrg() api.Organization {
	return apitest.NewOrganization()
}

func checkAccount() api.Account {
	return apitest.NewAccount()
}

// checkTestTheme creates a theme for testing
func checkTestTheme() *styles.Theme {
	return styles.NewTheme(true)
}

func keyMsg(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

// isCheckComplete checks if a step is complete by checking Next() doesn't return ErrNotReady
func isCheckComplete(s step.Step) bool {
	_, err := s.Next()
	return !errors.Is(err, step.ErrNotReady)
}

func TestCheckDatadogStep_Update(t *testing.T) {
	t.Parallel()
	org := checkOrg()
	account := checkAccount()

	t.Run("completes when datadog account exists", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccounts := &apitest.MockDatadogAccounts{
			HasAccountFunc: func(ctx context.Context, accountID string) (bool, error) {
				return true, nil
			},
			GetAccountFunc: func(ctx context.Context, accountID string) (*api.DatadogAccount, error) {
				return &api.DatadogAccount{ID: "dd-1", Name: "Production DD"}, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := datadog.NewCheckDatadogStep(context.Background(), checkTestTheme(), "admin", org, account, ddAccounts, apitest.NewMockWorkspaces(), preferencestest.NewMockPreferences(), apiClient, logger, nil)

		// Act: run init command
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !isCheckComplete(updated) {
			t.Error("expected step to complete when datadog exists")
		}
		// Verify Next() returns discovery step (not select region step)
		nextStep, err := updated.Next()
		if err != nil {
			t.Fatalf("expected no error from Next(), got %v", err)
		}
		if nextStep == nil {
			t.Fatal("expected Next() to return a step")
		}
	})

	t.Run("completes when no datadog account", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccounts := &apitest.MockDatadogAccounts{
			HasAccountFunc: func(ctx context.Context, accountID string) (bool, error) {
				return false, nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := datadog.NewCheckDatadogStep(context.Background(), checkTestTheme(), "admin", org, account, ddAccounts, apitest.NewMockWorkspaces(), preferencestest.NewMockPreferences(), apiClient, logger, nil)

		// Act: run init command
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !isCheckComplete(updated) {
			t.Error("expected step to complete when no datadog")
		}
		// Verify Next() returns select region step (needs setup)
		nextStep, err := updated.Next()
		if err != nil {
			t.Fatalf("expected no error from Next(), got %v", err)
		}
		if nextStep == nil {
			t.Fatal("expected Next() to return a step")
		}
	})

	t.Run("sets error state on failure", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ddAccounts := &apitest.MockDatadogAccounts{
			HasAccountFunc: func(ctx context.Context, accountID string) (bool, error) {
				return false, errors.New("API error")
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := datadog.NewCheckDatadogStep(context.Background(), checkTestTheme(), "admin", org, account, ddAccounts, apitest.NewMockWorkspaces(), preferencestest.NewMockPreferences(), apiClient, logger, nil)

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
		// When there's an error, Next() should return the error (not ErrNotReady)
		_, err := updated.Next()
		if err == nil {
			t.Error("expected Next() to return an error")
		}
		if errors.Is(err, step.ErrNotReady) {
			t.Error("expected Next() to return actual error, not ErrNotReady")
		}
	})

	t.Run("retries on r key when error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		attempts := 0
		ddAccounts := &apitest.MockDatadogAccounts{
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

		s := datadog.NewCheckDatadogStep(context.Background(), checkTestTheme(), "admin", org, account, ddAccounts, apitest.NewMockWorkspaces(), preferencestest.NewMockPreferences(), apiClient, logger, nil)

		// First attempt fails
		cmd := s.Init()
		updated := s
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		if !updated.HasError() {
			t.Fatal("expected error after first attempt")
		}

		// Act: press 'r' to retry
		updated, cmd = updated.Update(keyMsg('r'))
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if updated.HasError() {
			t.Error("expected error to be cleared after retry")
		}
		if !isCheckComplete(updated) {
			t.Error("expected step to complete after successful retry")
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})
}
