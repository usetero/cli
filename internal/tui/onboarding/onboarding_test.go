package onboarding

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/layouts/layoutstest"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/onboarding/step/steptest"
	"github.com/usetero/cli/internal/tui/onboarding/sync"
)

// onboardingTestTheme creates a theme for testing
func onboardingTestTheme() *styles.Theme {
	return styles.NewTheme(true)
}

func TestOnboarding_Update(t *testing.T) {
	// Note: Cannot use t.Parallel() on this test because one subtest uses t.Setenv
	t.Run("propagates error state from flow to layout immediately", func(t *testing.T) {
		// This test verifies the bug is fixed: when a step sets an error,
		// onboarding should propagate it to the layout in the same Update() call

		logger := logtest.New(t)
		layout := layoutstest.NewMockLayout()

		// Create a step that will have an error
		testStep := steptest.NewMockStep()

		// Create onboarding with the mock step
		onboarding := &Onboarding{
			flow:           step.NewFlow(testStep, logger),
			layout:         layout,
			ready:          true,
			logger:         logger,
			globalBindings: nil,
		}

		// Simulate an error message arriving at the step
		testErr := errors.New("test error")
		testStep.UpdateFunc = func(msg tea.Msg) (step.Step, tea.Cmd) {
			// Simulate the step receiving an error message and setting error state
			testStep.Err = testErr
			return testStep, nil
		}

		// Send a message to trigger the error
		onboarding.Update(tea.KeyPressMsg{})

		// BUG: Currently this fails because Update() calls layout.SetError(flow.Error())
		// BEFORE calling flow.Update(msg), so the error hasn't been set yet

		// After Update(), the layout should have received the error
		if layout.LastError == nil {
			t.Error("layout should have error after step sets error, got nil")
		} else if layout.LastError.Error() != testErr.Error() {
			t.Errorf("layout should have error after step sets error, got: %v, want: %v", layout.LastError, testErr)
		}
	})

	t.Run("completes when flow completes and extracts org and account", func(t *testing.T) {
		t.Parallel()
		logger := logtest.New(t)

		expectedOrg := api.Organization{ID: "org-123", Name: "Test Org"}
		expectedAccount := api.Account{ID: "acct-456", Name: "Test Account"}

		layout := layoutstest.NewMockLayout()

		// Create a sync step that holds the accumulated state
		syncStep := sync.New(context.Background(), onboardingTestTheme(), expectedOrg, expectedAccount, logger, nil)

		// Create a step that will transition to sync step
		testStep := steptest.NewMockStep()
		transitioned := false
		testStep.NextFunc = func() (step.Step, error) {
			if !transitioned {
				transitioned = true
				return syncStep, nil
			}
			return nil, step.ErrNotReady
		}

		onboarding := &Onboarding{
			flow:           step.NewFlow(testStep, logger),
			layout:         layout,
			ready:          true,
			logger:         logger,
			globalBindings: nil,
		}

		// First update triggers transition to sync step
		onboarding.Update(tea.KeyPressMsg{})

		// Send SyncReadyMsg - sync step handles it and completes
		onboarding.Update(powersync.SyncReadyMsg{})

		// Verify onboarding completed and extracted the org/account
		if !onboarding.IsComplete() {
			t.Error("expected onboarding to be complete")
		}

		if onboarding.Organization().ID != expectedOrg.ID {
			t.Errorf("expected org ID %s, got %s", expectedOrg.ID, onboarding.Organization().ID)
		}

		if onboarding.Account().ID != expectedAccount.ID {
			t.Errorf("expected account ID %s, got %s", expectedAccount.ID, onboarding.Account().ID)
		}
	})
}

func TestOnboarding_View(t *testing.T) {
	t.Parallel()
	t.Run("returns empty string when not ready", func(t *testing.T) {
		t.Parallel()
		logger := logtest.New(t)
		testStep := steptest.NewMockStep()
		layout := layoutstest.NewMockLayout()

		onboarding := &Onboarding{
			flow:           step.NewFlow(testStep, logger),
			layout:         layout,
			ready:          false, // Not ready
			logger:         logger,
			globalBindings: nil,
		}

		view := onboarding.View()
		if view != "" {
			t.Errorf("expected empty view when not ready, got: %s", view)
		}
	})
}
