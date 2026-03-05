package preferences

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domains/tenancy"
)

// Store persists preference snapshots.
type Store interface {
	Load(ctx context.Context) (Snapshot, error)
	Save(ctx context.Context, snapshot Snapshot) error
}

// PreferenceService is the domain contract for preference operations.
type PreferenceService interface {
	Snapshot(ctx context.Context) (Snapshot, error)
	SetRole(ctx context.Context, role Role) error
	SetOrganization(ctx context.Context, orgID tenancy.OrganizationID) error
	SetAccount(ctx context.Context, accountID tenancy.AccountID) error
	SetWorkspace(ctx context.Context, workspaceID tenancy.WorkspaceID) error
	SetScope(ctx context.Context, orgID tenancy.OrganizationID, accountID tenancy.AccountID, workspaceID tenancy.WorkspaceID) error
	ClearScope(ctx context.Context) error
}

// Service provides typed preference accessors for onboarding/runtime selection.
type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Snapshot(ctx context.Context) (Snapshot, error) {
	if s == nil || s.store == nil {
		return Snapshot{}, fmt.Errorf("preferences service is not initialized")
	}
	return s.store.Load(ctx)
}

func (s *Service) SetRole(ctx context.Context, role Role) error {
	if !role.Valid() {
		return fmt.Errorf("role is required")
	}
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Role = role
	})
}

func (s *Service) SetOrganization(ctx context.Context, orgID tenancy.OrganizationID) error {
	if orgID == "" {
		return fmt.Errorf("organization id is required")
	}
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Organization = orgID
		snapshot.Account = ""
		snapshot.Workspace = ""
	})
}

func (s *Service) SetAccount(ctx context.Context, accountID tenancy.AccountID) error {
	if accountID == "" {
		return fmt.Errorf("account id is required")
	}
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Account = accountID
		snapshot.Workspace = ""
	})
}

func (s *Service) SetWorkspace(ctx context.Context, workspaceID tenancy.WorkspaceID) error {
	if workspaceID == "" {
		return fmt.Errorf("workspace id is required")
	}
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Workspace = workspaceID
	})
}

func (s *Service) SetScope(ctx context.Context, orgID tenancy.OrganizationID, accountID tenancy.AccountID, workspaceID tenancy.WorkspaceID) error {
	if orgID == "" {
		return fmt.Errorf("organization id is required")
	}
	if accountID == "" {
		return fmt.Errorf("account id is required")
	}
	if workspaceID == "" {
		return fmt.Errorf("workspace id is required")
	}
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Organization = orgID
		snapshot.Account = accountID
		snapshot.Workspace = workspaceID
	})
}

func (s *Service) ClearScope(ctx context.Context) error {
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Organization = ""
		snapshot.Account = ""
		snapshot.Workspace = ""
	})
}

func (s *Service) update(ctx context.Context, mutate func(snapshot *Snapshot)) error {
	if s == nil || s.store == nil {
		return fmt.Errorf("preferences service is not initialized")
	}
	current, err := s.store.Load(ctx)
	if err != nil {
		return err
	}
	mutate(&current)
	return s.store.Save(ctx, current)
}
