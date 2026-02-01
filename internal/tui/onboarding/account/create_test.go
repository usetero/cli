package account_test

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
	"github.com/usetero/cli/internal/tui/onboarding/account"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/tuitest"
)

// createTestTheme creates a theme for testing
func createTestTheme() *styles.Theme {
	return styles.NewTheme(true)
}

func TestCreateStep_Update(t *testing.T) {
	t.Parallel()
	testOrg := api.Organization{ID: "org-1", Name: "Test Org"}

	t.Run("creates account on enter", func(t *testing.T) {
		t.Parallel()

		created := false
		accounts := &apitest.MockAccounts{
			CreateFunc: func(ctx context.Context, orgID string, name string) (*api.Account, error) {
				created = true
				return &api.Account{ID: "acc-new", Name: name}, nil
			},
		}
		prefs := &preferencestest.MockPreferences{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewCreateStep(context.Background(), createTestTheme(), "admin", testOrg, accounts, prefs, apiClient, logger, nil)

		// Type a name
		updated, _ := s.Update(tea.KeyPressMsg{Code: 'T', Text: "T"})
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 't', Text: "t"})

		// Press enter
		updated, cmd := updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		if !created {
			t.Error("expected account to be created")
		}
		_, err := updated.Next()
		if errors.Is(err, step.ErrNotReady) {
			t.Error("expected step to complete after creation")
		}
	})

	t.Run("does not submit empty input", func(t *testing.T) {
		t.Parallel()

		created := false
		accounts := &apitest.MockAccounts{
			CreateFunc: func(ctx context.Context, orgID string, name string) (*api.Account, error) {
				created = true
				return &api.Account{ID: "acc-new", Name: name}, nil
			},
		}
		prefs := &preferencestest.MockPreferences{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewCreateStep(context.Background(), createTestTheme(), "admin", testOrg, accounts, prefs, apiClient, logger, nil)

		// Press enter without typing anything
		updated, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		if created {
			t.Error("expected account NOT to be created with empty input")
		}
		_, err := updated.Next()
		if !errors.Is(err, step.ErrNotReady) {
			t.Error("expected step NOT to complete with empty input")
		}
	})

	t.Run("sets error state on failure", func(t *testing.T) {
		t.Parallel()

		accounts := &apitest.MockAccounts{
			CreateFunc: func(ctx context.Context, orgID string, name string) (*api.Account, error) {
				return nil, errors.New("API error")
			},
		}
		prefs := &preferencestest.MockPreferences{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewCreateStep(context.Background(), createTestTheme(), "admin", testOrg, accounts, prefs, apiClient, logger, nil)

		// Type a name
		updated, _ := s.Update(tea.KeyPressMsg{Code: 'T', Text: "T"})
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 't', Text: "t"})

		// Press enter
		updated, cmd := updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		if !updated.HasError() {
			t.Error("expected step to have error")
		}
	})

	t.Run("saves account ID to preferences", func(t *testing.T) {
		t.Parallel()

		savedID := ""
		accounts := &apitest.MockAccounts{
			CreateFunc: func(ctx context.Context, orgID string, name string) (*api.Account, error) {
				return &api.Account{ID: "acc-saved", Name: name}, nil
			},
		}
		prefs := &preferencestest.MockPreferences{
			SetDefaultAccountIDFunc: func(accountID string) error {
				savedID = accountID
				return nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewCreateStep(context.Background(), createTestTheme(), "admin", testOrg, accounts, prefs, apiClient, logger, nil)

		// Type and submit
		updated, _ := s.Update(tea.KeyPressMsg{Code: 'X', Text: "X"})
		updated, cmd := updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		if savedID != "acc-saved" {
			t.Errorf("expected account ID 'acc-saved' to be saved, got %q", savedID)
		}
	})

	t.Run("clears error on retry", func(t *testing.T) {
		t.Parallel()

		attempts := 0
		accounts := &apitest.MockAccounts{
			CreateFunc: func(ctx context.Context, orgID string, name string) (*api.Account, error) {
				attempts++
				if attempts == 1 {
					return nil, errors.New("first attempt fails")
				}
				return &api.Account{ID: "acc-retry", Name: name}, nil
			},
		}
		prefs := &preferencestest.MockPreferences{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewCreateStep(context.Background(), createTestTheme(), "admin", testOrg, accounts, prefs, apiClient, logger, nil)

		// Type and submit (first attempt - fails)
		updated, _ := s.Update(tea.KeyPressMsg{Code: 'X', Text: "X"})
		updated, cmd := updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		if !updated.HasError() {
			t.Fatal("expected error after first attempt")
		}

		// Press enter to retry (clears error)
		updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		if updated.HasError() {
			t.Error("expected error to be cleared on retry")
		}
	})
}
