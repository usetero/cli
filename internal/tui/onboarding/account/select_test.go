package account_test

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/remotelist"
	"github.com/usetero/cli/internal/tui/onboarding/account"
	"github.com/usetero/cli/internal/tui/onboarding/account/accounttest"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/tuitest"
)

// selectTestTheme creates a theme for testing
func selectTestTheme() *styles.Theme {
	return styles.NewTheme(true)
}

// isComplete checks if a step is complete by checking Next() doesn't return ErrNotReady
func isComplete(s step.Step) bool {
	_, err := s.Next()
	return err != step.ErrNotReady
}

func TestSelectStep_Update(t *testing.T) {
	t.Parallel()
	testOrg := api.Organization{ID: "org-1", Name: "Test Org"}

	t.Run("auto-selects when only one account exists", func(t *testing.T) {
		t.Parallel()
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

		s := account.NewSelectStep(context.Background(), selectTestTheme(), "admin", testOrg, lister, saver, apiClient, logger, nil)

		// Act: simulate load completing with one account
		items := []list.Item{account.AccountItem{ID: "acc-1", Name: "Production"}}
		updated, cmd := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Execute all commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !isComplete(updated) {
			t.Error("expected step to auto-select single account and complete")
		}
	})

	t.Run("auto-selects from saved preference", func(t *testing.T) {
		t.Parallel()
		// Arrange
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{
			GetDefaultAccountIDFunc: func() string {
				return "acc-2" // User has a saved preference
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewSelectStep(context.Background(), selectTestTheme(), "admin", testOrg, lister, saver, apiClient, logger, nil)

		// Act: simulate load completing with multiple accounts
		items := []list.Item{
			account.AccountItem{ID: "acc-1", Name: "Production"},
			account.AccountItem{ID: "acc-2", Name: "Staging"},
		}
		updated, cmd := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Execute all commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !isComplete(updated) {
			t.Error("expected step to auto-select from preference and complete")
		}
	})

	t.Run("auto-selects create when no accounts exist", func(t *testing.T) {
		t.Parallel()
		// Arrange
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewSelectStep(context.Background(), selectTestTheme(), "admin", testOrg, lister, saver, apiClient, logger, nil)

		// Act: simulate load completing with no accounts
		updated, _ := s.Update(remotelist.LoadResultMsg{Items: []list.Item{}, Err: nil})

		// Assert
		if !isComplete(updated) {
			t.Error("expected step to auto-select create and complete")
		}

		// Check that Next() returns a CreateStep (not nil)
		nextStep, err := updated.Next()
		if err != nil {
			t.Fatalf("expected no error from Next(), got %v", err)
		}
		if nextStep == nil {
			t.Error("expected Next() to return a step when create is selected")
		}
	})

	t.Run("requires manual selection when multiple accounts and no preference", func(t *testing.T) {
		t.Parallel()
		// Arrange
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{
			GetDefaultAccountIDFunc: func() string {
				return "" // No preference
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewSelectStep(context.Background(), selectTestTheme(), "admin", testOrg, lister, saver, apiClient, logger, nil)

		// Act: simulate load completing with multiple accounts
		items := []list.Item{
			account.AccountItem{ID: "acc-1", Name: "Production"},
			account.AccountItem{ID: "acc-2", Name: "Staging"},
		}
		updated, _ := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Assert: should NOT be complete - user must select
		if isComplete(updated) {
			t.Error("expected step to require manual selection")
		}
	})

	t.Run("saves preference when manually selecting account", func(t *testing.T) {
		t.Parallel()
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

		s := account.NewSelectStep(context.Background(), selectTestTheme(), "admin", testOrg, lister, saver, apiClient, logger, nil)

		// Load accounts first
		items := []list.Item{
			account.AccountItem{ID: "acc-1", Name: "Production"},
			account.AccountItem{ID: "acc-2", Name: "Staging"},
		}
		updated, _ := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Act: user presses enter to select
		updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Assert
		if savedID == "" {
			t.Error("expected account preference to be saved")
		}
		if !isComplete(updated) {
			t.Error("expected step to complete after selection")
		}
	})

	t.Run("selects create when user presses n", func(t *testing.T) {
		t.Parallel()
		// Arrange
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewSelectStep(context.Background(), selectTestTheme(), "admin", testOrg, lister, saver, apiClient, logger, nil)

		// Load accounts first
		items := []list.Item{
			account.AccountItem{ID: "acc-1", Name: "Production"},
		}
		updated, _ := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Act: user presses 'n' to create new
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})

		// Assert
		if !isComplete(updated) {
			t.Error("expected step to complete after pressing n")
		}

		// Check that Next() returns a CreateStep (not nil)
		nextStep, err := updated.Next()
		if err != nil {
			t.Fatalf("expected no error from Next(), got %v", err)
		}
		if nextStep == nil {
			t.Error("expected Next() to return a step when create is selected")
		}
	})
}

func TestSelectStep_Next(t *testing.T) {
	t.Parallel()
	testOrg := api.Organization{ID: "org-1", Name: "Test Org"}

	t.Run("returns ErrNotReady before selection", func(t *testing.T) {
		t.Parallel()
		// Arrange
		lister := &accounttest.MockAccountLister{}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewSelectStep(context.Background(), selectTestTheme(), "admin", testOrg, lister, saver, apiClient, logger, nil)

		// Assert
		_, err := s.Next()
		if err != step.ErrNotReady {
			t.Errorf("expected ErrNotReady before selection, got %v", err)
		}
	})
}
