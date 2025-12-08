package datadog_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/onboarding/datadog"
)

func TestSelectRegionStep_Update(t *testing.T) {
	t.Run("selects region on enter", func(t *testing.T) {
		// Arrange
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := datadog.NewSelectRegionStep("admin", "org-1", "acc-1", apiClient, logger, nil)

		// Act: press enter to select first region
		updated, _ := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Assert
		if !updated.IsComplete() {
			t.Error("expected step to complete after selection")
		}
		selectStep := updated.(*datadog.SelectRegionStep)
		if selectStep.SelectedRegion() == "" {
			t.Error("expected a region to be selected")
		}
	})

	t.Run("not complete before selection", func(t *testing.T) {
		// Arrange
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := datadog.NewSelectRegionStep("admin", "org-1", "acc-1", apiClient, logger, nil)

		// Assert
		if step.IsComplete() {
			t.Error("expected step NOT to be complete before selection")
		}
	})
}
