package organization_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/onboarding/organization"
	"github.com/usetero/cli/internal/tui/onboarding/organization/organizationtest"
	"github.com/usetero/cli/internal/tui/tuitest"
)

func TestCreateStep_Update(t *testing.T) {
	t.Run("creates organization on enter with valid input", func(t *testing.T) {
		// Arrange
		createCalled := false
		createdName := ""

		creator := &organizationtest.MockOrgCreator{
			CreateFunc: func(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error) {
				createCalled = true
				createdName = name
				return &api.OrganizationBootstrapResult{
					Organization: &api.Organization{
						ID:                   "org-123",
						Name:                 name,
						WorkosOrganizationID: "workos-123",
					},
					Account: &api.Account{ID: "acct-123"},
				}, nil
			},
		}
		refresher := &organizationtest.MockTokenRefresher{
			RefreshTokenWithOrganizationFunc: func(ctx context.Context, workosOrgID string) (string, error) {
				return "new-token", nil
			},
		}
		orgSaver := &organizationtest.MockDefaultOrgSaver{}
		accountSaver := &organizationtest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := organization.NewCreateStep("admin", creator, orgSaver, accountSaver, refresher, apiClient, logger, nil)

		// Type organization name
		for _, r := range "Acme Inc" {
			step.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}

		// Press enter to submit
		updated, cmd := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute create command
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}
		// Execute token refresh command
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Assert
		if !createCalled {
			t.Error("expected organization creator to be called")
		}
		if createdName != "Acme Inc" {
			t.Errorf("expected name 'Acme Inc', got %q", createdName)
		}
		if !updated.IsComplete() {
			t.Error("expected step to be complete after creation and token refresh")
		}
	})

	t.Run("does not submit on enter with empty input", func(t *testing.T) {
		// Arrange
		createCalled := false

		creator := &organizationtest.MockOrgCreator{
			CreateFunc: func(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error) {
				createCalled = true
				return nil, nil
			},
		}
		refresher := &organizationtest.MockTokenRefresher{}
		orgSaver := &organizationtest.MockDefaultOrgSaver{}
		accountSaver := &organizationtest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := organization.NewCreateStep("admin", creator, orgSaver, accountSaver, refresher, apiClient, logger, nil)

		// Press enter without typing anything
		updated, _ := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Assert
		if createCalled {
			t.Error("expected organization creator NOT to be called with empty input")
		}
		if updated.IsComplete() {
			t.Error("expected step to NOT be complete")
		}
	})

	t.Run("sets error state when creation fails", func(t *testing.T) {
		// Arrange
		creator := &organizationtest.MockOrgCreator{
			CreateFunc: func(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error) {
				return nil, errors.New("organization name already exists")
			},
		}
		refresher := &organizationtest.MockTokenRefresher{}
		orgSaver := &organizationtest.MockDefaultOrgSaver{}
		accountSaver := &organizationtest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := organization.NewCreateStep("admin", creator, orgSaver, accountSaver, refresher, apiClient, logger, nil)

		// Type and submit
		for _, r := range "Existing Org" {
			step.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		updated, cmd := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute create command (which fails)
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Assert
		if !updated.HasError() {
			t.Error("expected step to have error after creation fails")
		}
		if updated.IsComplete() {
			t.Error("expected step to NOT be complete when there's an error")
		}
	})

	t.Run("saves org and account IDs to preferences", func(t *testing.T) {
		// Arrange
		savedOrgID := ""
		savedAccountID := ""

		creator := &organizationtest.MockOrgCreator{
			CreateFunc: func(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error) {
				return &api.OrganizationBootstrapResult{
					Organization: &api.Organization{
						ID:                   "org-456",
						Name:                 name,
						WorkosOrganizationID: "workos-456",
					},
					Account: &api.Account{ID: "acct-789"},
				}, nil
			},
		}
		refresher := &organizationtest.MockTokenRefresher{
			RefreshTokenWithOrganizationFunc: func(ctx context.Context, workosOrgID string) (string, error) {
				return "new-token", nil
			},
		}
		orgSaver := &organizationtest.MockDefaultOrgSaver{
			SetDefaultOrgIDFunc: func(orgID string) error {
				savedOrgID = orgID
				return nil
			},
		}
		accountSaver := &organizationtest.MockDefaultAccountSaver{
			SetDefaultAccountIDFunc: func(accountID string) error {
				savedAccountID = accountID
				return nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := organization.NewCreateStep("admin", creator, orgSaver, accountSaver, refresher, apiClient, logger, nil)

		// Type and submit
		for _, r := range "New Org" {
			step.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		updated, cmd := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute commands
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Assert
		if savedOrgID != "org-456" {
			t.Errorf("expected org ID 'org-456' saved, got %q", savedOrgID)
		}
		if savedAccountID != "acct-789" {
			t.Errorf("expected account ID 'acct-789' saved, got %q", savedAccountID)
		}
	})

	t.Run("refreshes token and updates API client after creation", func(t *testing.T) {
		// Arrange
		refreshedOrgID := ""
		tokenSet := ""

		creator := &organizationtest.MockOrgCreator{
			CreateFunc: func(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error) {
				return &api.OrganizationBootstrapResult{
					Organization: &api.Organization{
						ID:                   "org-123",
						Name:                 name,
						WorkosOrganizationID: "workos-org-id",
					},
					Account: &api.Account{ID: "acct-123"},
				}, nil
			},
		}
		refresher := &organizationtest.MockTokenRefresher{
			RefreshTokenWithOrganizationFunc: func(ctx context.Context, workosOrgID string) (string, error) {
				refreshedOrgID = workosOrgID
				return "refreshed-token", nil
			},
		}
		orgSaver := &organizationtest.MockDefaultOrgSaver{}
		accountSaver := &organizationtest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{
			SetAccessTokenFunc: func(token string) {
				tokenSet = token
			},
		}
		logger := logtest.New(t)

		step := organization.NewCreateStep("admin", creator, orgSaver, accountSaver, refresher, apiClient, logger, nil)

		// Type and submit
		for _, r := range "Test Org" {
			step.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		updated, cmd := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Execute create command
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}
		// Execute token refresh command
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Assert
		if refreshedOrgID != "workos-org-id" {
			t.Errorf("expected token refresh with workos org ID 'workos-org-id', got %q", refreshedOrgID)
		}
		if tokenSet != "refreshed-token" {
			t.Errorf("expected API client token 'refreshed-token', got %q", tokenSet)
		}
	})

	t.Run("clears error and allows retry on enter", func(t *testing.T) {
		// Arrange
		callCount := 0

		creator := &organizationtest.MockOrgCreator{
			CreateFunc: func(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error) {
				callCount++
				if callCount == 1 {
					return nil, errors.New("temporary error")
				}
				return &api.OrganizationBootstrapResult{
					Organization: &api.Organization{ID: "org-123", Name: name, WorkosOrganizationID: ""},
					Account:      &api.Account{ID: "acct-123"},
				}, nil
			},
		}
		refresher := &organizationtest.MockTokenRefresher{}
		orgSaver := &organizationtest.MockDefaultOrgSaver{}
		accountSaver := &organizationtest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := organization.NewCreateStep("admin", creator, orgSaver, accountSaver, refresher, apiClient, logger, nil)

		// Type and submit (first attempt fails)
		for _, r := range "Retry Org" {
			step.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		updated, cmd := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Verify error state
		if !updated.HasError() {
			t.Fatal("expected error after first attempt")
		}

		// Press enter to retry (clears error)
		updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// Verify error cleared
		if updated.HasError() {
			t.Error("expected error to be cleared after retry")
		}
	})
}

func TestCreateStep_IsComplete(t *testing.T) {
	t.Run("returns false until created and token refreshed", func(t *testing.T) {
		// Arrange
		creator := &organizationtest.MockOrgCreator{
			CreateFunc: func(ctx context.Context, name string) (*api.OrganizationBootstrapResult, error) {
				return &api.OrganizationBootstrapResult{
					Organization: &api.Organization{ID: "org-123", Name: name, WorkosOrganizationID: "workos-123"},
					Account:      &api.Account{ID: "acct-123"},
				}, nil
			},
		}
		refresher := &organizationtest.MockTokenRefresher{
			RefreshTokenWithOrganizationFunc: func(ctx context.Context, workosOrgID string) (string, error) {
				return "new-token", nil
			},
		}
		orgSaver := &organizationtest.MockDefaultOrgSaver{}
		accountSaver := &organizationtest.MockDefaultAccountSaver{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		step := organization.NewCreateStep("admin", creator, orgSaver, accountSaver, refresher, apiClient, logger, nil)

		// Type and submit
		for _, r := range "Test" {
			step.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		updated, cmd := step.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		// After submit but before create completes
		if updated.IsComplete() {
			t.Error("should not be complete while creating")
		}

		// Execute create command (but not token refresh yet)
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// After create but before token refresh
		if updated.IsComplete() {
			t.Error("should not be complete before token refresh")
		}

		// Execute token refresh
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, cmd = updated.Update(msg)
		}

		// Now should be complete
		if !updated.IsComplete() {
			t.Error("should be complete after token refresh")
		}
	})
}
