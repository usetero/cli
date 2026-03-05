package onboardingtest

import (
	"context"

	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
)

type PreferenceService struct {
	SnapshotValue preferences.Snapshot
}

func (s *PreferenceService) Snapshot(context.Context) (preferences.Snapshot, error) {
	return s.SnapshotValue, nil
}
func (s *PreferenceService) SetRole(_ context.Context, role preferences.Role) error {
	s.SnapshotValue.Role = role
	return nil
}
func (s *PreferenceService) SetOrganization(_ context.Context, orgID tenancy.OrganizationID) error {
	s.SnapshotValue.Organization = orgID
	s.SnapshotValue.Account = ""
	s.SnapshotValue.Workspace = ""
	return nil
}
func (s *PreferenceService) SetAccount(_ context.Context, accountID tenancy.AccountID) error {
	s.SnapshotValue.Account = accountID
	s.SnapshotValue.Workspace = ""
	return nil
}
func (s *PreferenceService) SetWorkspace(_ context.Context, workspaceID tenancy.WorkspaceID) error {
	s.SnapshotValue.Workspace = workspaceID
	return nil
}
func (s *PreferenceService) SetScope(_ context.Context, orgID tenancy.OrganizationID, accountID tenancy.AccountID, workspaceID tenancy.WorkspaceID) error {
	s.SnapshotValue.Organization = orgID
	s.SnapshotValue.Account = accountID
	s.SnapshotValue.Workspace = workspaceID
	return nil
}
func (s *PreferenceService) ClearScope(_ context.Context) error {
	s.SnapshotValue.Organization = ""
	s.SnapshotValue.Account = ""
	s.SnapshotValue.Workspace = ""
	return nil
}
