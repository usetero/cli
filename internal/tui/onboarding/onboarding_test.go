package onboarding

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/layouts/layoutstest"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/onboarding/step/steptest"
)

// Mock preferences reader for testing
type mockPreferencesReader struct {
	getDefaultOrgIDFunc     func() string
	getDefaultAccountIDFunc func() string
}

func (m *mockPreferencesReader) GetDefaultOrgID() string {
	if m.getDefaultOrgIDFunc != nil {
		return m.getDefaultOrgIDFunc()
	}
	return ""
}

func (m *mockPreferencesReader) GetDefaultAccountID() string {
	if m.getDefaultAccountIDFunc != nil {
		return m.getDefaultAccountIDFunc()
	}
	return ""
}

func TestOnboarding_Update(t *testing.T) {
	t.Run("propagates error state from flow to layout immediately", func(t *testing.T) {
		// This test verifies the bug is fixed: when a step sets an error,
		// onboarding should propagate it to the layout in the same Update() call

		logger := logtest.New(t)
		prefs := &mockPreferencesReader{}
		layout := layoutstest.NewMockLayout()

		// Create a step that will have an error
		testStep := steptest.NewMockStep()

		// Create onboarding with the mock step
		onboarding := &Onboarding{
			flow:               step.NewFlow(testStep),
			layout:             layout,
			ready:              true,
			logger:             logger,
			preferencesService: prefs,
			globalBindings:     nil,
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

	t.Run("completes when flow completes and extracts org and account IDs", func(t *testing.T) {
		logger := logtest.New(t)

		expectedOrgID := "org-123"
		expectedAccountID := "acct-456"

		prefs := &mockPreferencesReader{
			getDefaultOrgIDFunc: func() string {
				return expectedOrgID
			},
			getDefaultAccountIDFunc: func() string {
				return expectedAccountID
			},
		}

		layout := layoutstest.NewMockLayout()

		// Create a completed step
		completedStep := steptest.NewMockStep()
		completedStep.IsCompleteFunc = func() bool {
			return true
		}

		onboarding := &Onboarding{
			flow:               step.NewFlow(completedStep),
			layout:             layout,
			ready:              true,
			logger:             logger,
			preferencesService: prefs,
			globalBindings:     nil,
		}

		// Update should extract IDs when complete
		onboarding.Update(tea.KeyPressMsg{})

		if !onboarding.IsComplete() {
			t.Error("onboarding should be complete when flow is complete")
		}

		if onboarding.OrganizationID() != expectedOrgID {
			t.Errorf("expected org ID %s, got %s", expectedOrgID, onboarding.OrganizationID())
		}

		if onboarding.AccountID() != expectedAccountID {
			t.Errorf("expected account ID %s, got %s", expectedAccountID, onboarding.AccountID())
		}
	})
}

func TestOnboarding_View(t *testing.T) {
	t.Run("returns empty string when not ready", func(t *testing.T) {
		logger := logtest.New(t)
		prefs := &mockPreferencesReader{}
		testStep := steptest.NewMockStep()
		layout := layoutstest.NewMockLayout()

		onboarding := &Onboarding{
			flow:               step.NewFlow(testStep),
			layout:             layout,
			ready:              false, // Not ready
			logger:             logger,
			preferencesService: prefs,
			globalBindings:     nil,
		}

		view := onboarding.View()
		if view != "" {
			t.Errorf("expected empty view when not ready, got: %s", view)
		}
	})
}
