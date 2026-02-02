package datadog_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/preferences/preferencestest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// selectRegionTestTheme creates a theme for testing
func selectRegionTestTheme() *styles.Theme {
	return styles.NewTheme(true)
}

// isRegionComplete checks if a step is complete by checking Next() doesn't return ErrNotReady
func isRegionComplete(s step.Step) bool {
	_, err := s.Next()
	return !errors.Is(err, step.ErrNotReady)
}

func TestSelectRegionStep_Update(t *testing.T) {
	t.Parallel()
	org := apitest.NewOrganization()
	account := apitest.NewAccount()

	t.Run("selects region on enter", func(t *testing.T) {
		t.Parallel()
		// Arrange
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := datadog.NewSelectRegionStep(context.Background(), selectRegionTestTheme(), "admin", org, account, apitest.NewMockWorkspaces(), preferencestest.NewMockPreferences(), apiClient, logger, nil)

		// Act: press enter to select first region
		updated, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Assert
		if !isRegionComplete(updated) {
			t.Error("expected step to complete after selection")
		}
		// Verify Next() returns the API key step
		nextStep, err := updated.Next()
		if err != nil {
			t.Fatalf("expected no error from Next(), got %v", err)
		}
		if nextStep == nil {
			t.Error("expected Next() to return API key step")
		}
	})

	t.Run("not complete before selection", func(t *testing.T) {
		t.Parallel()
		// Arrange
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := datadog.NewSelectRegionStep(context.Background(), selectRegionTestTheme(), "admin", org, account, apitest.NewMockWorkspaces(), preferencestest.NewMockPreferences(), apiClient, logger, nil)

		// Assert
		if isRegionComplete(s) {
			t.Error("expected step NOT to be complete before selection")
		}
	})
}
