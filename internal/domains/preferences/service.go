package preferences

import (
	"context"
)

// Store persists preference snapshots.
type Store interface {
	Load(ctx context.Context) (Snapshot, error)
	Save(ctx context.Context, snapshot Snapshot) error
}

// PreferenceService is the domain contract for preference operations.
type PreferenceService interface {
	Snapshot(ctx context.Context) (Snapshot, error)
	SetRole(ctx context.Context, selection RoleSelection) error
	SetOrganization(ctx context.Context, selection OrganizationSelection) error
	SetAccount(ctx context.Context, selection AccountSelection) error
	SetWorkspace(ctx context.Context, selection WorkspaceSelection) error
	SetScope(ctx context.Context, selection ScopeSelection) error
	ClearScope(ctx context.Context) error
}

// Service provides typed preference accessors for onboarding/runtime selection.
type Service struct {
	store Store
}

func NewService(store Store) *Service {
	if store == nil {
		panic("preferences service requires store")
	}
	return &Service{store: store}
}

func (s *Service) Snapshot(ctx context.Context) (Snapshot, error) {
	return s.store.Load(ctx)
}

func (s *Service) SetRole(ctx context.Context, selection RoleSelection) error {
	validated, err := selection.Validate()
	if err != nil {
		return err
	}
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Role = validated.Role
	})
}

func (s *Service) SetOrganization(ctx context.Context, selection OrganizationSelection) error {
	validated, err := selection.Validate()
	if err != nil {
		return err
	}
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Organization = validated.OrganizationID
		snapshot.Account = ""
		snapshot.Workspace = ""
	})
}

func (s *Service) SetAccount(ctx context.Context, selection AccountSelection) error {
	validated, err := selection.Validate()
	if err != nil {
		return err
	}
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Account = validated.AccountID
		snapshot.Workspace = ""
	})
}

func (s *Service) SetWorkspace(ctx context.Context, selection WorkspaceSelection) error {
	validated, err := selection.Validate()
	if err != nil {
		return err
	}
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Workspace = validated.WorkspaceID
	})
}

func (s *Service) SetScope(ctx context.Context, selection ScopeSelection) error {
	validated, err := selection.Validate()
	if err != nil {
		return err
	}
	return s.update(ctx, func(snapshot *Snapshot) {
		snapshot.Organization = validated.OrganizationID
		snapshot.Account = validated.AccountID
		snapshot.Workspace = validated.WorkspaceID
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
	current, err := s.store.Load(ctx)
	if err != nil {
		return err
	}
	mutate(&current)
	return s.store.Save(ctx, current)
}
