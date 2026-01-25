package account_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/onboarding/account"
	"github.com/usetero/cli/internal/tui/onboarding/account/accounttest"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/tuitest"
)

func TestCreateStep_Update(t *testing.T) {
	testOrg := api.Organization{ID: "org-1", Name: "Test Org"}

	t.Run("creates account on enter", func(t *testing.T) {
		// Arrange
		created := false
		creator := &accounttest.MockAccountCreator{
			CreateFunc: func(ctx context.Context, orgID string, name string) (*api.Account, error) {
				created = true
				return &api.Account{ID: "acc-new", Name: name}, nil
			},
		}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewCreateStep("admin", testOrg, creator, saver, apiClient, logger, nil)

		// Type a name
		updated, _ := s.Update(tea.KeyPressMsg{Code: 'T', Text: "T"})
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
		updated, _ = updated.Update(tea.KeyPressMsg{Code: 't', Text: "t"})

		// Act: press enter
		updated, cmd := updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !created {
			t.Error("expected account to be created")
		}
		// Check if step is complete by checking Next() doesn't return ErrNotReady
		_, err := updated.Next()
		if err == step.ErrNotReady {
			t.Error("expected step to complete after creation")
		}
	})

	t.Run("does not submit empty input", func(t *testing.T) {
		// Arrange
		created := false
		creator := &accounttest.MockAccountCreator{
			CreateFunc: func(ctx context.Context, orgID string, name string) (*api.Account, error) {
				created = true
				return &api.Account{ID: "acc-new", Name: name}, nil
			},
		}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewCreateStep("admin", testOrg, creator, saver, apiClient, logger, nil)

		// Act: press enter without typing anything
		updated, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if created {
			t.Error("expected account NOT to be created with empty input")
		}
		// Check if step is NOT complete by checking Next() returns ErrNotReady
		_, err := updated.Next()
		if err != step.ErrNotReady {
			t.Error("expected step NOT to complete with empty input")
		}
	})

	t.Run("sets error state on failure", func(t *testing.T) {
		// Arrange
		creator := &accounttest.MockAccountCreator{
			CreateFunc: func(ctx context.Context, orgID string, name string) (*api.Account, error) {
				return nil, errors.New("API error")
			},
		}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewCreateStep("admin", testOrg, creator, saver, apiClient, logger, nil)

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

		// Assert
		if !updated.HasError() {
			t.Error("expected step to have error")
		}
		// Check if step is NOT complete
		_, err := updated.Next()
		if err == nil || err == step.ErrNotReady {
			// If no error or just not ready, it means it didn't fail properly
			// Actually, when there's an error, Next() should return the error
			if err == step.ErrNotReady {
				t.Error("expected step NOT to complete on error")
			}
		}
	})

	t.Run("saves account ID to preferences", func(t *testing.T) {
		// Arrange
		savedID := ""
		creator := &accounttest.MockAccountCreator{
			CreateFunc: func(ctx context.Context, orgID string, name string) (*api.Account, error) {
				return &api.Account{ID: "acc-saved", Name: name}, nil
			},
		}
		saver := &accounttest.MockDefaultAccountSaver{
			SetDefaultAccountIDFunc: func(accountID string) error {
				savedID = accountID
				return nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewCreateStep("admin", testOrg, creator, saver, apiClient, logger, nil)

		// Type and submit
		updated, _ := s.Update(tea.KeyPressMsg{Code: 'X', Text: "X"})
		updated, cmd := updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if savedID != "acc-saved" {
			t.Errorf("expected account ID 'acc-saved' to be saved, got %q", savedID)
		}
	})

	t.Run("clears error on retry", func(t *testing.T) {
		// Arrange
		attempts := 0
		creator := &accounttest.MockAccountCreator{
			CreateFunc: func(ctx context.Context, orgID string, name string) (*api.Account, error) {
				attempts++
				if attempts == 1 {
					return nil, errors.New("first attempt fails")
				}
				return &api.Account{ID: "acc-retry", Name: name}, nil
			},
		}
		saver := &accounttest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		s := account.NewCreateStep("admin", testOrg, creator, saver, apiClient, logger, nil)

		// Type and submit (first attempt - fails)
		updated, _ := s.Update(tea.KeyPressMsg{Code: 'X', Text: "X"})
		updated, cmd := updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		if !updated.HasError() {
			t.Fatal("expected error after first attempt")
		}

		// Act: press enter to retry (clears error)
		updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Assert: error is cleared
		if updated.HasError() {
			t.Error("expected error to be cleared on retry")
		}
	})
}
