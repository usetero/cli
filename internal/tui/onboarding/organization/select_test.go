package organization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/auth/authtest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/preferences/preferencestest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/remotelist"
	"github.com/usetero/cli/internal/tui/onboarding/organization"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/tuitest"
)

// selectTestTheme creates a theme for testing
func selectTestTheme() *styles.Theme {
	return styles.NewTheme(true)
}

// isSelectComplete checks if a step is complete by checking Next() doesn't return ErrNotReady
func isSelectComplete(s step.Step) bool {
	_, err := s.Next()
	return !errors.Is(err, step.ErrNotReady)
}

func TestSelectStep_Update(t *testing.T) {
	t.Parallel()
	t.Run("auto-selects when only one organization exists", func(t *testing.T) {
		t.Parallel()
		// Arrange
		orgs := &apitest.MockOrganizations{
			ListFunc: func(ctx context.Context) ([]api.Organization, error) {
				return []api.Organization{
					{ID: "org-1", Name: "Acme Inc", WorkosOrganizationID: "workos-1"},
				}, nil
			},
		}
		prefs := &preferencestest.MockPreferences{}
		auth := &authtest.MockAuth{
			RefreshTokenWithOrganizationFunc: func(ctx context.Context, workosOrgID string) (string, error) {
				return "new-token", nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		s := organization.NewSelectStep(context.Background(), selectTestTheme(), "admin", orgs, workspaces, apiClient, prefs, auth, logger, nil)

		// Act: simulate load completing with one org
		items := []list.Item{organization.OrgItem{ID: "org-1", Name: "Acme Inc", WorkosOrganizationID: "workos-1"}}
		updated, cmd := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Execute all commands and process messages
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !isSelectComplete(updated) {
			t.Error("expected step to auto-select single org and complete")
		}
	})

	t.Run("auto-selects from saved preference", func(t *testing.T) {
		t.Parallel()
		// Arrange
		orgs := &apitest.MockOrganizations{}
		prefs := &preferencestest.MockPreferences{
			GetDefaultOrgIDFunc: func() string {
				return "org-2" // User has a saved preference
			},
		}
		auth := &authtest.MockAuth{
			RefreshTokenWithOrganizationFunc: func(ctx context.Context, workosOrgID string) (string, error) {
				return "new-token", nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		s := organization.NewSelectStep(context.Background(), selectTestTheme(), "admin", orgs, workspaces, apiClient, prefs, auth, logger, nil)

		// Act: simulate load completing with multiple orgs
		items := []list.Item{
			organization.OrgItem{ID: "org-1", Name: "Acme Inc", WorkosOrganizationID: "workos-1"},
			organization.OrgItem{ID: "org-2", Name: "Beta Corp", WorkosOrganizationID: "workos-2"},
		}
		updated, cmd := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Execute all commands and process messages
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !isSelectComplete(updated) {
			t.Error("expected step to auto-select from preference and complete")
		}
	})

	t.Run("auto-selects create when no organizations exist", func(t *testing.T) {
		t.Parallel()
		// Arrange
		orgs := &apitest.MockOrganizations{}
		prefs := &preferencestest.MockPreferences{}
		auth := &authtest.MockAuth{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		s := organization.NewSelectStep(context.Background(), selectTestTheme(), "admin", orgs, workspaces, apiClient, prefs, auth, logger, nil)

		// Act: simulate load completing with no orgs
		updated, _ := s.Update(remotelist.LoadResultMsg{Items: []list.Item{}, Err: nil})

		// Assert: should be complete (no token refresh needed for "create new")
		if !isSelectComplete(updated) {
			t.Error("expected step to auto-select create and complete")
		}

		// Verify it selected "create new" by checking Next() returns a step (CreateStep)
		nextStep, err := updated.Next()
		if err != nil {
			t.Fatalf("expected no error from Next(), got %v", err)
		}
		if nextStep == nil {
			t.Error("expected Next() to return CreateStep when create is selected")
		}
	})

	t.Run("requires manual selection when multiple orgs and no preference", func(t *testing.T) {
		t.Parallel()
		// Arrange
		orgs := &apitest.MockOrganizations{}
		prefs := &preferencestest.MockPreferences{
			GetDefaultOrgIDFunc: func() string {
				return "" // No preference
			},
		}
		auth := &authtest.MockAuth{}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		s := organization.NewSelectStep(context.Background(), selectTestTheme(), "admin", orgs, workspaces, apiClient, prefs, auth, logger, nil)

		// Act: simulate load completing with multiple orgs
		items := []list.Item{
			organization.OrgItem{ID: "org-1", Name: "Acme Inc", WorkosOrganizationID: "workos-1"},
			organization.OrgItem{ID: "org-2", Name: "Beta Corp", WorkosOrganizationID: "workos-2"},
		}
		updated, _ := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Assert: should NOT be complete - user must select
		if isSelectComplete(updated) {
			t.Error("expected step to require manual selection")
		}
	})

	t.Run("refreshes token after selecting organization", func(t *testing.T) {
		t.Parallel()
		// Arrange
		refreshCalled := false
		refreshedOrgID := ""

		orgs := &apitest.MockOrganizations{}
		prefs := &preferencestest.MockPreferences{}
		auth := &authtest.MockAuth{
			RefreshTokenWithOrganizationFunc: func(ctx context.Context, workosOrgID string) (string, error) {
				refreshCalled = true
				refreshedOrgID = workosOrgID
				return "new-token", nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		s := organization.NewSelectStep(context.Background(), selectTestTheme(), "admin", orgs, workspaces, apiClient, prefs, auth, logger, nil)

		// Simulate load completing with one org (triggers auto-select)
		items := []list.Item{organization.OrgItem{ID: "org-1", Name: "Acme Inc", WorkosOrganizationID: "workos-123"}}
		updated, cmd := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Execute all commands and process messages
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if !refreshCalled {
			t.Error("expected token refresh to be called")
		}
		if refreshedOrgID != "workos-123" {
			t.Errorf("expected refresh with workos org ID 'workos-123', got %q", refreshedOrgID)
		}
		if !isSelectComplete(updated) {
			t.Error("expected step to complete after token refresh")
		}
	})

	t.Run("updates API client with new token after refresh", func(t *testing.T) {
		t.Parallel()
		// Arrange
		tokenSet := ""

		orgs := &apitest.MockOrganizations{}
		prefs := &preferencestest.MockPreferences{}
		auth := &authtest.MockAuth{
			RefreshTokenWithOrganizationFunc: func(ctx context.Context, workosOrgID string) (string, error) {
				return "refreshed-access-token", nil
			},
		}
		apiClient := &apitest.MockClient{
			SetAccessTokenFunc: func(token string) {
				tokenSet = token
			},
		}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		s := organization.NewSelectStep(context.Background(), selectTestTheme(), "admin", orgs, workspaces, apiClient, prefs, auth, logger, nil)

		// Simulate load completing with one org (triggers auto-select)
		items := []list.Item{organization.OrgItem{ID: "org-1", Name: "Acme Inc", WorkosOrganizationID: "workos-1"}}
		updated, cmd := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Execute all commands and process messages
		for _, msg := range tuitest.DrainCmds(cmd) {
			updated, _ = updated.Update(msg)
		}

		// Assert
		if tokenSet != "refreshed-access-token" {
			t.Errorf("expected API client token to be set to 'refreshed-access-token', got %q", tokenSet)
		}
	})
}

func TestSelectStep_Next(t *testing.T) {
	t.Parallel()
	t.Run("returns ErrNotReady until token refreshed", func(t *testing.T) {
		t.Parallel()
		// Arrange
		orgs := &apitest.MockOrganizations{}
		prefs := &preferencestest.MockPreferences{}
		auth := &authtest.MockAuth{
			RefreshTokenWithOrganizationFunc: func(ctx context.Context, workosOrgID string) (string, error) {
				return "new-token", nil
			},
		}
		apiClient := &apitest.MockClient{}
		logger := logtest.New(t)

		workspaces := &apitest.MockWorkspaces{}
		s := organization.NewSelectStep(context.Background(), selectTestTheme(), "admin", orgs, workspaces, apiClient, prefs, auth, logger, nil)

		// Simulate load completing with one org
		items := []list.Item{organization.OrgItem{ID: "org-1", Name: "Acme Inc", WorkosOrganizationID: "workos-1"}}
		updated, _ := s.Update(remotelist.LoadResultMsg{Items: items, Err: nil})

		// Assert: should NOT be complete yet - token refresh command was returned but not executed
		if isSelectComplete(updated) {
			t.Error("expected step to NOT be complete before token refresh executes")
		}
	})
}
