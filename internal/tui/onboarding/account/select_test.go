package account_test

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/remotelist"
	"github.com/usetero/cli/internal/tui/onboarding/account"
	"github.com/usetero/cli/internal/tui/onboarding/account/accounttest"
	"github.com/usetero/cli/internal/tui/tuitest"
)

func TestSelectStep_Update(t *testing.T) {
	t.Run("auto-selects when only one account exists", func(t *testing.T) {
		// Arrange
		lister := &accounttest.MockAccountLister{
			ListFunc: func(ctx context.Context, orgID string) ([]api.Account, error) {
				return []api.Account{
					{ID: "acc-1", Name: "Production"},
				}, nil
			},
		}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := account.NewSelectStep("admin", "org-1", lister, saver, apiClient, logger, nil)

		// Act: simulate load completing with one account
		items := []list.Item{account.AccountItem{ID: "acc-1", Name: "Production"}}
		updated, cmd := step.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Execute all commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Assert
		if !updated.IsComplete() {
			t.Error("expected step to auto-select single account and complete")
		}
	})

	t.Run("auto-selects from saved preference", func(t *testing.T) {
		// Arrange
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{
			GetDefaultAccountIDFunc: func() string {
				return "acc-2" // User has a saved preference
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := account.NewSelectStep("admin", "org-1", lister, saver, apiClient, logger, nil)

		// Act: simulate load completing with multiple accounts
		items := []list.Item{
			account.AccountItem{ID: "acc-1", Name: "Production"},
			account.AccountItem{ID: "acc-2", Name: "Staging"},
		}
		updated, cmd := step.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Execute all commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Assert
		if !updated.IsComplete() {
			t.Error("expected step to auto-select from preference and complete")
		}
	})

	t.Run("auto-selects create when no accounts exist", func(t *testing.T) {
		// Arrange
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := account.NewSelectStep("admin", "org-1", lister, saver, apiClient, logger, nil)

		// Act: simulate load completing with no accounts
		updated, _ := step.Update(remotelist.LoadResultMsg{Items: []list.Item{}, Err: nil})

		// Assert
		if !updated.IsComplete() {
			t.Error("expected step to auto-select create and complete")
		}

		selectStep := updated.(*account.SelectStep)
		if !selectStep.IsCreateSelected() {
			t.Error("expected create to be selected when no accounts exist")
		}
	})

	t.Run("requires manual selection when multiple accounts and no preference", func(t *testing.T) {
		// Arrange
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{
			GetDefaultAccountIDFunc: func() string {
				return "" // No preference
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := account.NewSelectStep("admin", "org-1", lister, saver, apiClient, logger, nil)

		// Act: simulate load completing with multiple accounts
		items := []list.Item{
			account.AccountItem{ID: "acc-1", Name: "Production"},
			account.AccountItem{ID: "acc-2", Name: "Staging"},
		}
		updated, _ := step.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Assert: should NOT be complete - user must select
		if updated.IsComplete() {
			t.Error("expected step to require manual selection")
		}
	})

	t.Run("saves preference when manually selecting account", func(t *testing.T) {
		// Arrange
		savedID := ""
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{
			SetDefaultAccountIDFunc: func(accountID string) error {
				savedID = accountID
				return nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := account.NewSelectStep("admin", "org-1", lister, saver, apiClient, logger, nil)

		// Load accounts first
		items := []list.Item{
			account.AccountItem{ID: "acc-1", Name: "Production"},
			account.AccountItem{ID: "acc-2", Name: "Staging"},
		}
		updated, _ := step.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Act: user presses enter to select
		updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Assert
		if savedID == "" {
			t.Error("expected account preference to be saved")
		}
		if !updated.IsComplete() {
			t.Error("expected step to complete after selection")
		}
	})

	t.Run("selects create when user presses n", func(t *testing.T) {
		// Arrange
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := account.NewSelectStep("admin", "org-1", lister, saver, apiClient, logger, nil)

		// Load accounts first
		items := []list.Item{
			account.AccountItem{ID: "acc-1", Name: "Production"},
		}
		updated, _ := step.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Act: user presses 'n' to create new
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})

		// Assert
		if !updated.IsComplete() {
			t.Error("expected step to complete after pressing n")
		}

		selectStep := updated.(*account.SelectStep)
		if !selectStep.IsCreateSelected() {
			t.Error("expected create to be selected")
		}
	})
}

func TestSelectStep_IsComplete(t *testing.T) {
	t.Run("returns false before selection", func(t *testing.T) {
		// Arrange
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := account.NewSelectStep("admin", "org-1", lister, saver, apiClient, logger, nil)

		// Assert
		if step.IsComplete() {
			t.Error("expected step to NOT be complete before any selection")
		}
	})
}
